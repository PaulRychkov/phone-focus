# Сбор данных об использовании приложений — UsageStatsManager

**Вывод:** приложение полностью заменяет Digital Wellbeing штатным AOSP-API `android.app.usage.UsageStatsManager`. GMS не нужен. Приложение ванпласа («Цифровой комфорт») читать не надо и нельзя (см. [02-oneplus-coloros.md](02-oneplus-coloros.md)).

## Разрешение
- `PACKAGE_USAGE_STATS` — это **special access**, НЕ выдаётся `requestPermissions()`. В манифесте:
  `<uses-permission android:name="android.permission.PACKAGE_USAGE_STATS" tools:ignore="ProtectedPermissions"/>` (иначе lint-ошибка ProtectedPermissions).
- Пользователь включает вручную: Settings → Спец. доступ → «Доступ к истории использования». Deep-link: `Intent(Settings.ACTION_USAGE_ACCESS_SETTINGS)`.
- Проверка выдачи в рантайме: `AppOpsManager.unsafeCheckOpNoThrow(OPSTR_GET_USAGE_STATS, Process.myUid(), packageName) == MODE_ALLOWED`; при `MODE_DEFAULT` — fallback `checkCallingOrSelfPermission(PACKAGE_USAGE_STATS)`. Колбэка нет — перепроверять в `onResume()` после возврата из настроек.

## Какой метод
- `queryEvents(begin, end) → UsageEvents` — **основной**: дискретный поток событий, единственный способ восстановить точную ленту, текущее приложение и произвольное окно.
- `queryUsageStats` / `queryAndAggregateUsageStats` — грубые бакеты (INTERVAL_DAILY…), лагают до rollover/flush, для «последних минут» не годятся.
- События foreground: `ACTIVITY_RESUMED=1` / `ACTIVITY_PAUSED=2` (API 29; старые имена `MOVE_TO_FOREGROUND/BACKGROUND` — те же int, код по `eventType==1/2` версионно-устойчив). `ACTIVITY_STOPPED=23` — более твёрдый сигнал ухода, чем PAUSED (PAUSED срабатывает и на диалоги-оверлеи).

## Текущее приложение и окно 10 минут
- Текущее foreground = пакет последнего `ACTIVITY_RESUMED`, за которым не последовал свой `ACTIVITY_PAUSED`, в коротком трейлинг-окне (~60с). Кэшировать последнее известное — при слишком коротком окне queryEvents может вернуть пусто в середине сессии.
- Rolling-10-мин на пакет: `queryEvents(now-600000, now)`, пары RESUMED→PAUSED/STOPPED, клэмп границ сессии к окну. Если сессия началась ДО окна (нет RESUMED внутри) — засеять открытую сессию на `windowStart`, иначе недосчёт.

## Латентность и версии
- `queryEvents` практически realtime (AOSP мерджит in-memory + disk). Системный лаг ~2.5с — для 10-мин окна не важен.
- **Android 16 / API 36 — никаких ограничений на usage stats** (обе страницы behavior-changes молчат; проверено). Ключевой исторический сдвиг — только депрекейт MOVE_TO_* в API 29.
- UsageStatsManager — pull-only, колбэка на смену foreground нет → нужен периодический сэмплер (см. [03-background-scheduling.md](03-background-scheduling.md)).

## Gotchas
- Запрашивать окно, сдвинутое в прошлое; текущий момент часто пуст.
- `getTotalTimeInForeground()` считает целые сессии, к 10-мин границе не клэмпится — использовать queryEvents + ручной клэмп.
- OEM-hibernation может отозвать special access после долгого простоя — перепроверять на resume.
- Данные per-user/per-profile.

## Источники
- https://developer.android.com/reference/android/app/usage/UsageStatsManager
- https://developer.android.com/reference/android/app/usage/UsageEvents.Event
- https://developer.android.com/reference/android/app/AppOpsManager
- https://developer.android.com/about/versions/16/behavior-changes-16 , /behavior-changes-all
- https://medium.com/@quiro91/show-app-usage-with-usagestatsmanager-d47294537dab
- https://github.com/Cap-go/capacitor-android-usagestatsmanager
