# Интеграция с motivation-bot и экосистемой

Изучен код `C:\Users\pavel\motivation-bot` и контракты `productivity-ecosystem\docs`.

## Как бот принимает события сейчас
- Единый generic-путь: Kafka-консьюмер (`bot-core-group`, топики `tasks.events`/`pomodoro.events`/`bot.agent-responses`/`bot.tg-sent`) → `HandleKafka` ([dispatcher.go](../../motivation-bot/internal/core/app/dispatcher.go)) свитчит по топику → `ingestExternalEvent` разбирает CloudEvents `Envelope` (`{specversion,id,source,type,subject,time,datacontenttype,data}`), требует непустые `id/source/type`, вставляет в `inbox_events` с `ON CONFLICT (source,event_id) DO NOTHING`. **Код ингеста source-агностичен: новый источник = новый топик в свитче, больше здесь ничего.**
- `ProcessInbox` → `logic.Match` — **детерминированный матчер, только МЕНЯЕТ outcome уже открытых интервенций**, ничего не создаёт. Новый source='phone' будет лежать в inbox инертно, пока код не создаст интервенцию.

## Кто создаёт интервенции (важно!)
Три создателя: (1) `scheduler.go` (morning_plan/evening_review по времени), (2) `deadline.go` (pull tasks REST → deadline_reminder), (3) агент через тул `send_ping` → `CreateAgentIntervention` (жёстко ограничен `free_ping`/`question`). Все через `createIntervention` (гарды: пауза, quiet hours, бюджет free_ping, suppression) → `requestAgentText` в `bot.agent-requests`.

Две LLM-роли: полный **ReAct-агент** (`handleUpdate`) запускается ТОЛЬКО на входящее сообщение пользователя (`bot.tg-updates`), у него тулы; и **tool-less генератор текста** (`handleAgentRequest`) — только пишет body уже созданной интервенции. **Сейчас ни одно ВХОДЯЩЕЕ СОБЫТИЕ не запускает автономное решение агента «стоит ли ткнуть».** → нудж от события телефона = core детерминированно создаёт touch, LLM его лишь озвучивает.

## Транспорт: выбран вариант B — HTTPS-ingest на core :8083
- **A (телефон → Kafka напрямую на tt-k3s:9094) — сломан и небезопасен**: advertised listener возвращает `localhost:9094` (классическая ловушка — телефон полезет к себе), PLAINTEXT без auth, «not internet-hardened», а cloudflared туннелит только HTTP/HTTPS, не TCP Kafka.
- **C (отдельный HTTP→Kafka сервис)** — лишний компонент, против минимализма.
- **B**: телефон POST-ит CloudEvents-JSON на `POST /ingest/v1/phone-usage` (bearer-токен `BOT_PHONE_INGEST_TOKEN`), хендлер продюсит в новый топик `phone.events`, который тот же `bot-core-group` уже generic-роутит в inbox. Переиспользует весь dedup/dispatch, сохраняет наблюдаемость в kafka-ui. (Чуть легче — писать сразу в `inbox_events` из хендлера, но продюс в топик консистентнее контракту.)
- ⚠️ Сетевой путь до :8083 из РФ — см. [04-build-install-egress.md](04-build-install-egress.md) (БЛОКЕР Cloudflare). Открытый вопрос: проброшен ли :8083 наружу через туннель (память: cloudflared в k3s через tun2socks).

## Контракт CloudEvent (snake_case, CloudEvents 1.0)
Envelope: `specversion "1.0"`, `id` = UUID на снапшот (**переиспользовать при ретрае** — иначе ON CONFLICT не сдедупит), `source "phone"`, `type "phone.usage.snapshot"` (+ опц. лёгкий `phone.foreground.changed`), `subject` = device_id, `time` = window_end (RFC3339 UTC).
`data{}`: `device_id`, `chat_id` (int, привязка к профилю), `window_start`, `window_end`, `foreground_app`, `foreground_package`, `foreground_category` (work|productivity|communication|social|video|games|reading|other), `foreground_since` (когда текущее приложение вышло на передний план — даёт непрерывный dwell между снапшотами), `screen_on` (bool), `apps` (массив `{package,label,category,foreground_seconds}`), `distracting_seconds` (int, предпосчёт клиента), `changed_since_last` (bool, skip-if-unchanged), `seq` (int, детект пропусков).

## Точный список изменений в motivation-bot (phase-2/3)
**MODIFY:**
1. `internal/kafkax/kafkax.go` — `const TopicPhoneEvents = "phone.events"`.
2. `internal/core/models/models.go` — `const SourcePhone = "phone"`.
3. `internal/core/app/dispatcher.go` — в `HandleKafka` добавить `case kafkax.TopicPhoneEvents:` → `ingestExternalEvent` + `ProcessInbox` + `evaluateDistraction`.
4. `cmd/core-service/main.go` — добавить топик в `topics` и в `consumer.RequireSuccess`.
5. `internal/core/handler/handler.go` — `POST /ingest/v1/phone-usage` за bearer-middleware; расширить интерфейс `Service`.
6. `internal/config/config.go` — `BOT_PHONE_INGEST_TOKEN`, порог `DistractionThresholdMin`.
7. `internal/core/app/app.go` — метод ингеста (валидирует CloudEvent, продюсит в `phone.events`) + `RecentPhoneUsage(ctx, chatID, minutes)` по inbox `WHERE source='phone'`.
8. `prompts/dialog.md` — описать тул `get_recent_phone_usage` и политику нуджа.

**ADD:**
9. `internal/core/app/distraction.go` — `evaluateDistraction(ctx, snapshot)`: порог по dwell/категории → `createIntervention` (free_ping, Context `{reason:"distraction", app, category, minutes}`). Событийно, без нового goroutine.

**Опц. (тул для ReAct):** `get_recent_phone_usage` в `internal/agent/tools/core.go` + метод в `coreclient/client.go` + `GET /internal/v1/phone-usage/recent`.

**Схема:** `inbox_events.source` — свободный текст с `UNIQUE(source,event_id)` → source='phone' **без миграции**. НО `intervention_kind` — Postgres ENUM: отдельный `kind='distraction'` потребует миграцию + texts.go + prompt. **Чтобы не мигрировать — переиспользовать `free_ping`** (расходует бюджет 3/день) или **`question`** (без бюджета), причину нести в Context.

## Gotchas / открытые вопросы
- Матчер только меняет открытые интервенции — phone-событие само по себе инертно.
- ReAct-агент бежит только на сообщение юзера → «агент решает» = core решает детерминированно, LLM озвучивает; полноценное агентное решение на каждое событие — заметно больший механизм.
- Ретеншн inbox — 90 дней; частые снапшоты раздуют `inbox_events` → коалесить на клиенте (`changed_since_last`), либо короче ретеншн для source='phone', либо хранить только дельты.
- Параметры политики: порог непрерывного dwell, какие категории «отвлекающие», кулдаун между нуджами, гейтить ли на активную focus-сессию pomodoro (данные доступны через inbox/MCP).
- Привязка chat: система фактически single-user (один профиль) — нести `chat_id` в payload явно (предложено) или маппить device_id на единственный профиль.
