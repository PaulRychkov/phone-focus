# Надёжный фоновый сэмплер каждые ~10 минут (Android 15/16, ColorOS, без GMS)

**Вывод:** слоёная само-восстанавливающаяся схема. Ни один одиночный механизм не выживает на ColorOS.

## Почему не то, что кажется очевидным
- **PeriodicWorkManager НЕ может 10 минут** — жёсткий минимум `repeatInterval` = 15 мин (клэмпится молча).
- **dataSync / mediaProcessing FGS** — кап 6 ч / 24 ч на Android 15+ → `onTimeout()` → надо `stopSelf()` иначе `RemoteServiceException`. Таймер сбрасывается только когда пользователь открывает приложение. Не годится для 24/7.
- **shortService FGS** — ~3 мин. Не годится.

## Рекомендуемая схема
1. **Хост — foreground service типа `specialUse`** (`android:foregroundServiceType="specialUse"` + `<property android:name="android.app.PROPERTY_SPECIAL_USE_FGS_SUBTYPE" .../>`; permission `FOREGROUND_SERVICE` + `FOREGROUND_SERVICE_SPECIAL_USE`). У `specialUse` **нет** таймаута 6ч/3мин и он **разрешён при boot** (в отличие от dataSync). Play-декларация specialUse касается только публикации в Play — **для sideload не применяется**; манифест-декларация обязательна (её проверяет OS). Нотификация — на канале IMPORTANCE_MIN/LOW, ongoing.
2. **Heartbeat — одноразовый `AlarmManager.setExactAndAllowWhileIdle(~10мин)`**, переустанавливаемый в `BroadcastReceiver` на каждый тик. В ресивере: снять сэмпл **инлайн** (на Android 16 НЕ плодить quota-bound Jobs — джобы из FGS теперь подчиняются своим квотам), убедиться что FGS жив (перезапустить если ColorOS убил), поставить следующий алярм. **Firing exact alarm — явное исключение** из запрета Android 12+ на старт FGS из фона → алярм может воскресить сервис.
   - Без промпта на разрешение: `setAndAllowWhileIdle()` (неточный, пол ~9 мин в Doze) — для 10-мин цели джиттер приемлем.
3. **Разрешения/исключения (первый запуск):**
   - `POST_NOTIFICATIONS` (Android 13+).
   - `SCHEDULE_EXACT_ALARM` — по умолчанию **deny** на targetSdk 33+; `ACTION_REQUEST_SCHEDULE_EXACT_ALARM` + гард `canScheduleExactAlarms()`. Либо `USE_EXACT_ALARM` (авто-грант, Play-ограничение не касается sideload). Либо вовсе неточный вариант без разрешения.
   - `REQUEST_IGNORE_BATTERY_OPTIMIZATIONS` (`ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`) — **двойная роль**: снимает throttling Doze/App-Standby И сам является исключением для старта FGS из фона.
4. **Автозапуск:** `BOOT_COMPLETED` / `LOCKED_BOOT_COMPLETED` / `MY_PACKAGE_REPLACED` ресиверы перезапускают specialUse FGS (specialUse разрешён при boot на Android 15). **+ OEM autostart-онбординг** (см. [02](02-oneplus-coloros.md)) — единственная защита от ColorOS-киллов.

## Doze / App Standby
- «Rare» bucket = ~10 мин job-runtime в сутки + жёсткий throttling алярмов. Battery-exempt + FGS держат приложение вне карающих bucket'ов; `allow-while-idle` алярмы и FGS — escape hatch из Doze.

## Ничего не требует GMS.

## Источники
- https://developer.android.com/develop/background-work/services/fgs/timeout
- https://developer.android.com/develop/background-work/services/fgs/restrictions-bg-start
- https://developer.android.com/about/versions/15/changes/foreground-service-types
- https://developer.android.com/develop/background-work/services/alarms
- https://developer.android.com/topic/performance/appstandby , /training/monitoring-device-state/doze-standby
- https://developer.android.com/develop/background-work/services/fgs/changes (Android 16)
- https://dontkillmyapp.com/oppo , /oneplus
