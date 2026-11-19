# NovaMessenger

An independent cross-platform messaging application with its own backend.
This project is not connected to Telegram servers and does not use Telegram accounts.

## Planned technology

- Backend: Go, Gin, WebSocket
- Data: PostgreSQL and Redis
- Client: Flutter / Dart
- Deployment: Docker Compose, VPS or a secure tunnel for temporary testing

## Repository layout

- `backend/` — Go API and WebSocket server
- `mobile/` — Flutter application (Android, Windows and Web targets)
- `docker/` — deployment-related files
- `docs/` — project documentation

## Initial setup

1. Install Git and Docker Desktop.
2. Copy `.env.example` to `.env`.
3. Replace every `REPLACE_WITH_...` value with a long, unique secret.
4. Never commit or share `.env`.

The backend source and Dockerfile will be added in the next project package.
Do not run `docker compose up` until those files have been supplied.

## Security

- Never publish `.env`, database passwords, JWT secrets or private keys.
- Developer/test features must be authorized by the backend.
- Do not expose PostgreSQL or Redis ports publicly.
- Use HTTPS/WSS and firewall rules on a public VPS.

## License

MIT. Add a `LICENSE` file before publishing if you choose this license.
