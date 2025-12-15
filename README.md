# User Registration API (Go, Fiber, MongoDB)

Simple RESTful service for managing users with JWT authentication, MongoDB persistence, and minimal dependencies.

## Features
- Register and login users with bcrypt-hashed passwords.
- JWT (HS256) auth middleware protecting `/api/**`.
- CRUD: list, get, update, delete users.
- MongoDB storage via official driver.
- HTTP logging middleware (method, path, duration).
- Background task every 10s logging user count.
- JSON startup logs.

## Prerequisites
- Go 1.24+
- MongoDB running and reachable (default: `mongodb://localhost:27017`, DB: `userdb`).

## Configuration
Edit `config/config.yml`:
```yaml
server:
  port: ":8080"

mongo:
  uri: "mongodb://localhost:27017"
  db_name: "userdb"

app:
  jwt_secret: "change_this_in_prod"
```

## Run
```sh
go run .
```
Visit `http://localhost:8080/health` for a quick check. Adjust the port in config if needed.

## API
- `POST /register` — create user. Body: `{"name":"Alice","email":"alice@example.com","password":"secret"}`.
- `POST /login` — returns `{"token":"<jwt>"}`. Body: `{"email":"alice@example.com","password":"secret"}`.
- Authenticated (Bearer token):
  - `GET /api/users` — list users.
  - `GET /api/users/:id` — get by ID.
  - `PUT /api/users/:id` — update name/email. Body: `{"name":"New","email":"new@example.com"}`.
  - `DELETE /api/users/:id` — delete.

## Logging
- Structured JSON at startup for routes and server start.
- Request logging via middleware: `METHOD PATH DURATION`.
- Background goroutine logs total user count every 10s.

## Testing
```sh
go test ./...
```
