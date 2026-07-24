# OnePlus 13T PKX110 / ColorOS 16 CN — специфика

## Устройство и ОС
- **OnePlus 13T PKX110** — китайский вариант (глобальный близнец — OnePlus 13S / CPH2723). В Китае работает **ColorOS**, не OxygenOS.
- Билд **`16.0.8.300 CN01B60P01`** → **ColorOS 16 на базе Android 16 (API 36)**, `CN` = китайский ROM. Телефон вышел 24.04.2025 на ColorOS 15 / Android 15, обновлён до ColorOS 16 (раскатка в Китае с 30.10.2025; для PKX110 подтверждены билды 16.x).
- Точное соответствие именно `16.0.8.300` ↔ Android 16 не нашлось в чейнджлоге пословно (confidence: подтверждено косвенно). **Проверить на устройстве**: Settings → About → Android version.

## Приложение «Цифровой комфорт и забота о семье» v16.6.6
- Это OEM-реализация ColorOS, пакет **`com.coloros.digitalwellbeing`** (НЕ гугловский `com.google.android.apps.wellbeing`, которого на CN ROM нет).
- Данные — приватная SQLite в `/data/data/com.coloros.digitalwellbeing/`, **без exported ContentProvider и публичного API** → сторонним приложением не читается без root (песочница per-UID). Форензика подтверждает: доступ только при извлечении из `/data/data` (root/физический дамп).
- **Вывод: не пытаться читать его БД.** Тот же исходный сигнал даёт `UsageStatsManager` — это и есть санкционированный путь (см. [01-usagestats.md](01-usagestats.md)).

## Нет Google Play Services
- CN ColorOS идёт без GMS (нет Play Store/Services). FCM без VPN не достукивается до Google → **не строить планирование/пуш на FCM или GMS**. `UsageStatsManager`, `WorkManager`, `AlarmManager`, FGS — чистый AOSP, от GMS не зависят.

## Агрессивное убийство фона (главный риск надёжности)
Кода-фикса НЕТ, только пользовательский онбординг (dontkillmyapp.com/oneplus, /oppo):
1. **Autostart** — включить приложение в «Startup manager» / «Allow Auto Start-up» в security-app (`com.coloros.safecenter` / `com.coloros.phonemanager`).
2. **Батарея** — App info → Battery → «Allow background activity»; Battery optimization → «Don't optimize»; в «Advanced optimization» ОТКЛЮЧИТЬ «Deep optimization/Adaptive Battery» (главный киллер) и «Sleep standby optimization» (режет сеть/джобы).
3. **Залочить карточку в Recents** (long-press → замок) — на OnePlus это ещё и не даёт «Don't optimize» самопроизвольно откатываться (известное поведение: молча возвращает через ~сутки → перепроверять и ре-промптить).
4. Постоянная foreground-нотификация у сервиса.

## Установка self-signed APK на CN ColorOS
- Per-source «install unknown apps» (разрешить конкретный установщик).
- Сканер безопасности (`com.oplus.appdetail` + `com.coloros.securityguard` + `com.coloros.phonemanager`) на КАЖДЫЙ sideload → «risk found» → «install anyway» может дёрнуть «Verify your identity» (OPPO/HeyTap аккаунт + SMS OTP, который может не прийти на не-CN номер; воркэраунд — вход по email/WeChat).
- Обход для dev-устройства: `adb shell pm uninstall --user 0 com.oplus.appdetail` (и securityguard/phonemanager) — убирает сканер; или просто ставить через `adb install` (минует установщик-гейт).
- Dev options: Settings → About → Version → 7 тапов по Build number; затем USB debugging + Wireless debugging (может понадобиться выключить «permission monitoring» для pairing).

## Источники
- https://www.gizmochina.com/2025/04/24/oneplus-13t-launched-in-china-with-snapdragon-8-elite-and-6260mah-battery/
- https://www.huaweicentral.com/coloros-16-devices-oppo-oneplus/
- https://community.oppo.com/thread/2072903929544310784
- https://android-developers.googleblog.com/2025/06/android-16-is-here.html
- https://dontkillmyapp.com/oneplus , https://dontkillmyapp.com/oppo
- https://thedroidwin.com/how-to-disable-app-verification-on-coloros/
- https://thebinaryhick.blog/2020/02/22/walking-the-android-timeline-using-androids-digital-wellbeing-to-timeline-android-activity/
- https://xdaforums.com/t/are-there-any-users-in-china-using-coloros-16-ive-heard-that-fcm-has-been-restricted.4764879/
