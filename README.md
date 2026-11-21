# FlexGram Private

Приватный офлайн-клон Telegram для личного использования, контента и розыгрышей.  
Все данные хранятся **только локально** (localStorage). Никаких серверов Telegram.

## Стек

- **React 18 + Vite + TypeScript**
- **Zustand** (состояние + persist)
- **Capacitor 6** → Android APK
- **Electron** → Windows / Linux / macOS
- **GitHub Actions** → автоматическая сборка

## Возможности

- UI в стиле Telegram (тёмная/светлая тема, папки, stories, пузыри, reply, long-press меню)
- Голосовые сообщения (симуляция), подарки, Stars, Premium
- **Director Mode** (Режиссёр переписки): двойной клик по шапке чата → писать от имени собеседника, системные сообщения, накрутка просмотров
- Локальный кошелёк (TON / USDT / Notcoin)
- Имитация звонков
- Секретные чаты
- Экспорт/импорт бэкапа JSON
- Деcoy-аккаунт (демо-режим)

## Быстрый старт

```bash
npm install
npm run dev
```

Открой http://localhost:5173

### Desktop (Electron)

```bash
npm run electron:dev          # разработка
npm run electron:build:win    # .exe
```

### Android (Capacitor)

```bash
npm run build
npx cap add android
npx cap sync android
npx cap open android
# В Android Studio: Build → Build APK
```

## GitHub Actions

При пуше в `main` или ручном запуске workflow:

1. Собирает web-дистрибутив
2. Готовит Android-проект (APK — при наличии SDK в runners)
3. Собирает Windows `.exe` через electron-builder
4. Создаёт draft Release с артефактами

## Структура

```
src/
  store/useStore.ts     # всё состояние + local persist
  screens/              # Auth, Main, Chat, Settings, Gifts, Call
  types/                # TypeScript-модели
electron/main.cjs       # Electron entry
capacitor.config.ts
.github/workflows/build.yml
```

## Важно

Это **не** клиент Telegram и **не** подключается к каким-либо серверам.  
Предназначен для локальных сценариев, демо, контента и безопасных розыгрышей.

MIT License
