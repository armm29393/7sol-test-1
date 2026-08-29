# User management REST API — Echo + MongoDB + JWT auth.

## Stack

- Go, [Echo](https://echo.labstack.com/) (HTTP framework)
- MongoDB (persistence)
- JWT (`golang-jwt/jwt/v5`) for auth, `bcrypt` for password hashing
- `go-playground/validator` for request validation

## Project layout

```
cmd/api                  entrypoint (main.go)
internal/config          env config loader
internal/connector       external clients (mongo.go)
internal/auth            password hashing + JWT sign/parse
internal/domain/user     User model, request payloads, domain errors
internal/repository/user Repository interface + Mongo implementation
internal/usecase/user    business logic
internal/handler         router, middleware, validator
internal/handler/user    REST handlers + route registration
collection/              Bruno API collection
```

## Setup

Requirements: Go 1.27+, Docker (for MongoDB).

1. Copy env file:
   ```bash
   cp .env.example .env.local
   ```
2. Fill in `.env.local` (see [Environment variables](#environment-variables) below).
3. Start MongoDB:
   ```bash
   make mongo
   ```

## Run

```bash
make dev
```

`make dev` loads `.env.local` (or `.env`) and runs `go run ./cmd/api`. Server starts on `PORT` (default `8080`).

Other useful targets: `make build`, `make test`, `make cover`, `make check` (fmt+vet+test). Run `make help` for the full list.

### Run with Docker

```bash
make docker-up      # docker compose up -d --build (api + mongo)
make docker-down
```

Docker compose reads env vars from `.env.prod`, not `.env.local`.

## Environment variables

| Var          | Description                          | Example                          |
|--------------|---------------------------------------|-----------------------------------|
| `ENV`        | environment name (log only)           | `development`                    |
| `PORT`       | HTTP port                             | `8080`                           |
| `MONGO_URI`  | Mongo connection string               | `mongodb://localhost:27017`      |
| `MONGO_DB`   | Mongo database name                   | `userdb`                         |
| `JWT_SECRET` | HMAC secret for signing/verifying JWT | (see below)                      |

## Generating a JWT secret

```bash
make secret
```

Runs `openssl rand -base64 32`. Copy the output into `JWT_SECRET` in your env file.

## Using JWT

- `POST /register` and `POST /login` are public.
- `/login` returns `{ "token": "<jwt>" }`. Token is signed HS256, expires in 24h (`internal/auth/jwt.go`).
- All `/users/*` routes require `Authorization: Bearer <token>`. Missing/invalid token → `401`.
- The token's `sub` claim is the user ID; `JWTMiddleware` decodes it and stores it as `userID` in the request context (not currently used to scope queries — see Assumptions).

## Sample requests / responses

### Register

```
POST /register
Content-Type: application/json

{
  "name": "Arm",
  "email": "arm@mail.com",
  "password": "12121212"
}
```

`201 Created`
```json
{
  "id": "6a926fd252c5e88c01851487",
  "name": "Arm",
  "email": "arm@mail.com",
  "created_at": "2026-08-29T12:36:18.998252+07:00"
}
```

`409 Conflict` (email already registered)
```json
{ "message": "email already registered" }
```

### Login

```
POST /login
Content-Type: application/json

{
  "email": "arm@mail.com",
  "password": "12121212"
}
```

`200 OK`
```json
{ "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

`401 Unauthorized` (bad credentials) → `{ "message": "invalid email or password" }`

### Authenticated user routes

All require `Authorization: Bearer <token>`.

```
GET    /users        -> 200, [ { user }, ... ]
GET    /users/:id    -> 200, { user }  |  404 { "message": "user not found" }
PUT    /users/:id    { "name": "...", "email": "..." } -> 200, { user }
DELETE /users/:id    -> 204
```

User object shape (password never serialized):
```json
{
  "id": "6a926fd252c5e88c01851487",
  "name": "Arm",
  "email": "arm@mail.com",
  "created_at": "2026-08-29T12:36:18.998252+07:00"
}
```

A ready-to-import Bruno collection is in `collection/` (`collection/opencollection.yml`).

## Assumptions & design decisions

- **No authorization/ownership check**: any authenticated user can read/update/delete any user by ID — `userID` from the JWT is decoded but not used to restrict access. Acceptable for a test project; would need an ownership or role check for real use.
- **Repository pattern**: `internal/repository/user` defines the `Repository` interface next to its Mongo implementation, so the usecase layer depends only on the interface, not Mongo directly.
- **Errors as sentinel values** (`domain/user/errors.go`): usecase returns domain errors (`ErrNotFound`, `ErrEmailExists`, `ErrInvalidCreds`); the HTTP layer maps them to status codes in one place (`errorResponse` in `resthandler.go`) so Mongo-specific errors never leak to clients.
- **Config has defaults, not validation**: `config.Load()` falls back to dev defaults (e.g. `JWT_SECRET=change-me`) instead of failing if env vars are missing — convenient for local dev, but means a misconfigured prod deploy fails silently rather than refusing to start.
- **Background goroutine in `main.go`**: logs total user count every 10s as a demonstration of graceful shutdown wiring (context cancellation on `SIGINT`/`SIGTERM`), not a required feature.
- **Module-owned routes**: each domain module (currently just `user`) registers its own routes via `RegisterRoutes`, keeping `internal/handler/router.go` free of per-route knowledge — new modules plug in without touching the router.
