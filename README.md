# FlexGram — Private Messenger

Независимый кроссплатформенный мессенджер (приватный клон Telegram).
Собственный бэкенд, никакой связи с серверами Telegram.

## Стек

- **Backend**: Go + Gin + WebSocket (gorilla/websocket)
- **DB**: PostgreSQL + Redis
- **Client**: Flutter (Android / Windows / Web)

## Возможности

- Регистрация / логин по username + пароль
- Личные чаты, поиск пользователей по @username
- Real-time сообщения через WebSocket
- Профиль с градиентной шапкой, Premium, галочка верификации
- Магазин коллекционных подарков + баланс Stars
- **God Mode** (Developer Panel) — выдача себе Stars, Premium, подарков, любых username

## Быстрый старт (сервер)

```bash
# 1. Клонировать
git clone https://github.com/arii0012/flexgram-private.git
cd flexgram-private

# 2. Создать .env
cp .env.example .env
# Отредактируйте .env — поставьте длинные пароли и JWT_SECRET

# 3. Запустить всё одной командой
docker compose up -d --build

# Сервер будет на http://localhost:8080
```

Проверка:
```bash
curl http://localhost:8080/
```

## Flutter-клиент

```bash
cd mobile
flutter pub get

# Android (эмулятор)
# В lib/services/auth.dart уже стоит baseUrl = 'http://10.0.2.2:8080'
flutter run

# Windows
# Раскомментируйте baseUrl = 'http://localhost:8080' в auth.dart
flutter run -d windows

# Сборка APK
flutter build apk --release
# Файл: build/app/outputs/flutter-apk/app-release.apk

# Сборка Windows EXE
flutter build windows --release
```

### Важно для двух устройств

Если хотите, чтобы телефон и ПК писали друг другу:

1. Узнайте IP вашего компьютера в локальной сети (например `192.168.1.50`).
2. В `mobile/lib/services/auth.dart` поставьте:
   ```dart
   static const String baseUrl = 'http://192.168.1.50:8080';
   ```
3. В `docker-compose.yml` порты уже проброшены.
4. Убедитесь, что файрвол разрешает входящие на 8080.

## God Mode

В приложении: меню → **Developer Panel**.

Можно:
- Накинуть себе любое количество Stars
- Выдать Premium + галочку
- Сменить username на короткий (3–4 символа)
- Выдать себе любой подарок с кастомным номером

## Структура

```
├── backend/                 # Go сервер
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── handlers/        # auth, chats, messages, gifts, god
│   │   ├── websocket/       # real-time hub
│   │   ├── models/
│   │   └── database/migrations/
│   └── Dockerfile
├── mobile/                  # Flutter клиент
│   └── lib/
│       ├── screens/         # login, home, chat, profile, gifts, god
│       └── services/        # api + auth
├── docker-compose.yml
└── README.md
```

## API (кратко)

| Method | Path | Описание |
|--------|------|----------|
| POST | /api/auth/register | Регистрация |
| POST | /api/auth/login | Вход |
| GET  | /api/users/me | Профиль |
| GET  | /api/users/search?q= | Поиск |
| POST | /api/chats | Создать чат |
| GET  | /api/chats | Список чатов |
| POST | /api/chats/:id/messages | Отправить сообщение |
| GET  | /api/chats/:id/messages | История |
| GET  | /api/gifts/shop | Магазин |
| POST | /api/gifts/buy | Купить подарок |
| POST | /api/god/stars | Накинуть Stars |
| POST | /api/god/premium | Premium + verified |
| GET  | /ws?token=... | WebSocket |

## Лицензия

MIT

## Автосборка на GitHub

При каждом пуше в `main` GitHub Actions автоматически:

1. Собирает Go-сервер (Linux + Windows)
2. Собирает Android APK
3. Создаёт **Release** с готовыми файлами для скачивания

### Как пользоваться

1. Запушь код в репозиторий
2. Открой вкладку **Actions** — увидишь сборку
3. После успеха открой вкладку **Releases** — там будут:
   - `flexgram-android.apk`
   - `flexgram-server-linux`
   - `flexgram-server-windows.exe`

Можно также запустить сборку вручную: **Actions → Build FlexGram → Run workflow**.
