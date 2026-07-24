# Сборка/подпись/установка APK + сетевой egress из РФ

## Сборка на Windows без Android Studio (работает полноценно)
- **JDK**: минимум 17, рекомендую Temurin **21** (JAVA_HOME).
- **Android SDK CLI**: скачать `cmdline-tools`, затем `sdkmanager "platform-tools" "platforms;android-36" "build-tools;36.0.0"`, `yes | sdkmanager --licenses`. ANDROID_HOME.
- Сборка: `gradlew.bat assembleRelease` (wrapper сам качает Gradle). Android Studio не нужен — это обычный CI-рецепт.
- **Стек середины 2026**: AGP **9.x** (минимум для API 36 — 8.9.0; 9.0.1 — янв 2026), Gradle **9.x** (через wrapper), Kotlin **2.3.x**. ⚠️ AGP 9.0 имеет встроенную поддержку Kotlin — НЕ применять `org.jetbrains.kotlin.android` вручную (конфликт плагинов). Точные минорки (AGP ~9.3 / Gradle ~9.5 к июлю 2026) — перепроверить при сборке на developer.android.com/build/releases.
- `compileSdk=36`, `targetSdk=36`, `minSdk` 29–34 (для одного устройства без разницы).

## Подпись
- Debug — авто-подпись `debug.keystore`, `assembleDebug` сразу ставится.
- Release — self-signed keystore: `keytool -genkeypair -keystore rel.jks -alias k -keyalg RSA -keysize 2048 -validity 10000`. Прописать `signingConfigs {}` в `build.gradle.kts` → Gradle сам делает zipalign + apksigner (v1/v2/v3). Для sideload self-signed достаточно.
- ⚠️ Не путать: keystore подписи APK ≠ TLS-сертификат сервера.

## Установка
- USB: dev options + USB debugging → `adb install -r app.apk`.
- Wireless (Android 11+): на телефоне Wireless debugging → «Pair device with pairing code»; `adb pair <IP>:<pairPort>` (код) → `adb connect <IP>:<connectPort>` (порт connect ≠ pair, меняется каждую сессию) → `adb install`.
- ColorOS-гейты установки — см. [02-oneplus-coloros.md](02-oneplus-coloros.md).

## ⚠️ Сетевой egress из РФ (2026) — БЛОКЕР ДЛЯ ПУТИ «ТЕЛЕФОН→БОТ»
- **Cloudflare edge режется**: с июня 2025 все крупные RU-ISP (вкл. мобильных **MTS, Megafon, Beeline, Rostelecom, MGTS**) пропускают только первые ~16 КБ ответа Cloudflare, дальше drop + потеря пакетов. Публичный hostname `cloudflared` резолвится в edge-IP Cloudflare → **троттлится так же, как любой сайт на CF, независимо от того, где стоит VM**. Это и есть текущий туннель к tt-k3s.
- **Telegram Bot API**: `api.telegram.org` ~95% заблокирован без VPN (апрель 2026, эскалирует). Обфускация Telegram апреля 2026 защищает клиент MTProto, НЕ server-to-server HTTPS к Bot API.
- **Рекомендация phase-1**: телефон POST-ит на **кастомный HTTPS-ingest на ПРЯМОМ origin-IP** (VM/VPS + Let's Encrypt), НЕ через Cloudflare edge и НЕ напрямую в Telegram. Если и VM, и телефон в РФ — RU→RU напрямую работает. Payload держать маленьким. TLS к публично-доверенному сертификату → без `network_security_config`/cleartext-конфига. Android 17 включит Certificate Transparency по умолчанию → предпочесть Let's Encrypt (CT-логируемый CA), не приватный CA.
- Если оставить Cloudflare Tunnel в v1 — держать запрос+ответ сильно < ~16 КБ, один запрос на соединение, retries/backoff, best-effort.
- **Открытые вопросы к Павлу** (решить до старта): где реально достижим tt-k3s (есть ли публичный IPv4 + inbound 443, или он только за cloudflared / за CGNAT)? Где физически VM (в РФ?)? Будет ли телефон ходить через VPN/прокси (тогда CF/Telegram снова доступны)? На каком операторе телефон?

## Источники
- https://developer.android.com/about/versions/16/setup-sdk
- https://developer.android.com/build/jdks , https://developer.android.com/build/releases/gradle-plugin
- https://developer.android.com/tools/adb , https://developer.android.com/studio/publish/app-signing
- https://blog.cloudflare.com/russian-internet-users-are-unable-to-access-the-open-internet/
- https://ppc.land/russian-isps-throttle-cloudflare-traffic-to-16kb/
- https://meduza.io/en/news/2026/04/10/telegram-blocking-rate-in-russia-reaches-95
- https://developer.android.com/privacy-and-security/security-config
