# Сборка

Всё окружение поднято локально, в одну папку `C:\Users\pavel\android-tools`. В систему ничего не ставилось, права администратора не нужны, Android Studio не нужен. Чтобы убрать всё — удалить эту папку.

## Что где лежит

| Что | Путь | Версия |
|---|---|---|
| JDK | `C:\Users\pavel\android-tools\jdk-extract\jdk-21.0.11+10` | Temurin 21 LTS |
| Android SDK | `C:\Users\pavel\android-tools\sdk` | платформа 36, build-tools 36, platform-tools |
| NDK | `C:\Users\pavel\android-tools\sdk\ndk\30.0.15729638` | нужен для сборки Go под Android |
| Gradle | `C:\Users\pavel\android-tools\gradle-8.11.1` | 8.11.1 |
| gomobile | `C:\Users\pavel\go\bin\gomobile.exe` | golang.org/x/mobile |
| Ключ подписи | `C:\Users\pavel\android-tools\phone-focus-release.jks` | самоподписанный, 10000 дней |

Пути к SDK и пароли ключа лежат в `android/local.properties` и `android/keystore.properties` — оба в `.gitignore`.

## Шаг 1. Собрать Go-ядро в библиотеку

```powershell
$env:JAVA_HOME="C:\Users\pavel\android-tools\jdk-extract\jdk-21.0.11+10"
$env:ANDROID_HOME="C:\Users\pavel\android-tools\sdk"
$env:ANDROID_NDK_HOME="C:\Users\pavel\android-tools\sdk\ndk\30.0.15729638"
$env:PATH="$env:JAVA_HOME\bin;C:\Users\pavel\go\bin;$env:PATH"
cd C:\Users\pavel\phone-focus
gomobile bind -target=android -androidapi 24 -o android/app/libs/focus.aar ./mobile
```

Библиотека собирается под четыре архитектуры (arm64-v8a, armeabi-v7a, x86, x86_64).

Важно: в `go.mod` есть строка `tool golang.org/x/mobile/cmd/gobind` — без неё `gomobile bind` отказывается работать, а `go mod tidy` вычищает зависимость.

## Шаг 2. Собрать интерфейс

```powershell
cd C:\Users\pavel\phone-focus\frontend
npm install
npm run build
```

Vite кладёт готовые файлы в `android\app\src\main\assets`, откуда их показывает WebView.

## Шаг 3. Собрать APK

```powershell
& "C:\Users\pavel\android-tools\gradle-8.11.1\bin\gradle.bat" -p "C:\Users\pavel\phone-focus\android" assembleRelease
```

Готовый файл: `android\app\build\outputs\apk\release\app-release.apk`.

## Шаг 3. Поставить на телефон

```powershell
$adb="C:\Users\pavel\android-tools\sdk\platform-tools\adb.exe"
& $adb devices
& $adb install -r "C:\Users\pavel\phone-focus\android\app\build\outputs\apk\release\app-release.apk"
```

По Wi-Fi (Android 11+): на телефоне «Для разработчиков» → «Отладка по Wi-Fi» → «Подключить с помощью кода», затем `adb pair <ip>:<порт-сопряжения>`, ввести код, потом `adb connect <ip>:<порт-подключения>`.

Установка через `adb` минует проверку безопасности ColorOS, которая ругается на приложения не из магазина.

## Проверить Go-часть

```powershell
cd C:\Users\pavel\phone-focus
go test ./...
go vet ./...
```

## Стек

Go 1.25, gomobile bind; Kotlin 2.1.10, AGP 8.9.1, Gradle 8.11.1; `compileSdk`/`targetSdk` 36, `minSdk` 24. Внешних зависимостей у Go-части одна — `github.com/google/uuid`; у Kotlin-части нет ни одной, кроме самой библиотеки `focus.aar`.
