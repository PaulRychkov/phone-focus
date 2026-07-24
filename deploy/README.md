# Доставка данных с телефона боту

Телефон не может достучаться до бота напрямую: VM с ботом — за NAT Hyper-V, снаружи её видно только через туннель, а российский IP VM режется Cloudflare. Поэтому цепочка такая:

```
Телефон (AdGuard-VPN вкл, выход Рига)
   → https://phone-focus.tensor-talks.ru   (Cloudflare edge, из Риги без троттлинга)
   → cloudflared на ПК (свой туннель phone-focus, отдельный от TensorTalks)
   → SSH-проброс 127.0.0.1:18083 → VM:8083   (прямой путь хост→VM по docker-портам сломан, идём через SSH)
   → motivation-bot core :8083  POST /ingest/v1/phone-usage  (bearer-токен)
   → Kafka phone.events → инбокс бота → решение о напоминании → Telegram
```

## Что где настроено

- **Туннель** `phone-focus` (id `1ef57cdf-…`) — отдельный, туннель `tensor-talks-new` не тронут. Конфиг: `C:\Users\pavel\.cloudflared\phone-focus.yml` — **protocol http2** (QUIC за AdGuard-VPN мигал ошибкой 530/1033; http2 через TCP/443 стабилен). DNS `phone-focus.tensor-talks.ru` → этот туннель (поддомен вынужденно под tensor-talks.ru — единственная зона в аккаунте; при своём домене перецепить одной командой `cloudflared tunnel route dns`).
- **SSH-проброс + туннель** держит скрипт `deploy/phone-focus-relay.ps1`, автозапуск — задача планировщика `PhoneFocusRelay` (при входе в систему, с рестартом при падении).
- **Токен приёма** `BOT_PHONE_INGEST_TOKEN` — в `~/motivation-bot/docker-compose.override.yml` на VM (core-service). Ротация: поменять там, `sudo docker compose up -d core-service`, вписать новый в приложение.
- **Бот** развёрнут на VM (`sudo docker compose up -d --build core-service agent-service`): топик `phone.events`, эндпоинт `/ingest/v1/phone-usage`, `logic.EvaluateDistraction`, интервенция `question`.

## Настройка приложения

На экране Focus:
- Адрес сервера: `https://phone-focus.tensor-talks.ru`
- Токен доступа: значение `BOT_PHONE_INGEST_TOKEN` с VM
- chat_id: `1986746872`
- На телефоне AdGuard-VPN должен быть **включён** (иначе Cloudflare режется на РФ-IP).

## Зависимости и слабые места

- VM крутится в Hyper-V на этом ПК → бот доступен только когда ПК включён (та же доступность, что и у остального контура).
- IP VM в подсети Hyper-V ротируется; проброс идёт по имени `tt-k3s` (hosts). При смене IP — обновить hosts (скрипт `sync-and-bootstrap.ps1`).
- Нужен включённый AdGuard-VPN на ПК (egress cloudflared в Ригу) и на телефоне.
- Прямой путь хост→VM по docker-портам (8083/8089/9094) сломан на этой машине (Docker+Hyper-V): TCP-рукопожатие проходит, данные — нет; SSH работает, поэтому проброс через него.
