# Paccy — Backend

REST API for [Paccy](https://github.com/paccy-io/paccy-frontend), a college application
workspace: users, college applications, essays, extracurriculars, academics and
calendar events.

## Stack

- Go, [Fiber](https://gofiber.io)
- MongoDB
- JWT auth, Argon2id password hashing
- Email via Resend or Gmail SMTP, with per-project and per-user quotas
- Optional Discord webhook for error, startup and sign-up alerts

## Layout

```
cmd/            entrypoint
api/            HTTP handlers, middleware, routes
internal/       feature packages (auth, user, college, essay, …), each with
                models, DTOs, repository and service
pkg/            shared helpers (logger, errors, responses, hashing)
```

## Getting started

Requires Go 1.24+ and a running MongoDB.

```bash
cp .env.example .env   # then fill in real values
go run ./cmd/main.go
```

The API serves on `http://localhost:3000/api`.

`JWT_SECRET_KEY` must be set to a long random value whenever
`APP_ENVIRONMENT=production`; the server refuses to start otherwise.

## Docker

```bash
docker build -t paccy-backend .
docker run --env-file .env -p 3000:3000 paccy-backend
```

## License

[MIT](LICENSE)
