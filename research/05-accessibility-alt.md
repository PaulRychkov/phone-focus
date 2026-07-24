# UsageStatsManager (polling) vs AccessibilityService (realtime)

**Вывод для phase-1: только UsageStatsManager. Accessibility — не сейчас.**

## Сравнение
| | UsageStatsManager (queryEvents) | AccessibilityService |
|---|---|---|
| Латентность | секунды (~2.5с сист. лаг + интервал опроса) | мгновенно (sub-100ms, `TYPE_WINDOW_STATE_CHANGED` → `getPackageName()`) |
| Грант | один тумблер «Usage access» | `BIND_ACCESSIBILITY_SERVICE` + ручной тумблер в Settings→Спец. возможности, не программно |
| Постоянный сервис | не нужен | нужен привилегированный bound-сервис |
| Переживает ребут | да | нужен ре-энейбл |
| На ColorOS | ок | **ColorOS специально ОТКЛЮЧАЕТ accessibility при screen-off** → ручной ре-энейбл; permission «revoked when app closes» |
| Android 13+ sideload | — | «Restricted Settings» гасит тумблер → обход через App info → «Allow restricted settings» |
| Батарея | легче (батчинг queryEvents) | event-driven дёшев в простое, но постоянный bound-сервис + keep-alive FGS = постоянный расход |

## Почему для 10-мин снапшота хватает UsageStats
Смысл 10-мин окна — грубая гранулярность; ~2.5с лаг queryEvents абсолютно неважен против окна в 600с. Один грант, без always-on привилегированного сервиса, переживает ребут, и — решающе на ColorOS — **не отключается молча**, как accessibility на каждый screen-off.

## Когда добавлять Accessibility (phase-4, опционально)
Только если понадобятся **интервенции в реальном времени** («ткнуть в секунду открытия отвлекающего приложения»). Тогда принять: постоянный FGS + полный ColorOS keep-alive чеклист + одноразовый «Allow restricted settings» + логика мониторинга «сервис ещё включён?» и ре-промпта. Прагматичный гибрид: UsageStats как надёжный бэкбон отчётности + accessibility поверх как низколатентный триггер с деградацией на (чуть запоздалые) UsageStats-нуджи.

## Gotchas
- queryEvents на текущий момент возвращает пусто — запрашивать окно в прошлом.
- ColorOS отключает accessibility (не пауза — именно off), нужно активно детектить состояние.
- Android 12+ auto-reset/hibernation отзывает разрешения после ~90 дней простоя (для ежедневного приложения не риск).

## Источники
- https://developer.android.com/reference/android/accessibilityservice/AccessibilityService
- https://developer.android.com/guide/topics/ui/accessibility/service
- https://dontkillmyapp.com/oppo
- https://www.androidpolice.com/android-13-blocks-accessibility-services-sideloaded-apps/
- http://www.timelesssky.com/blog/technical-details-of-app-gatekeeper-version-1-1
