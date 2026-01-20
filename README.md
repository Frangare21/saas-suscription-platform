# SaaS Subscription Platform (mini Stripe/Paddle)

Plataforma SaaS orientada a empresas que cobran **suscripciones**. El objetivo del repo es construir un conjunto de microservicios (estilo Stripe/Paddle/MercadoPago para SaaS) para:

- registrar usuarios
- autenticar y emitir JWT
- gestionar suscripciones y facturación (en progreso)
- (a futuro) pagos, notificaciones, webhooks, etc.

> Estado actual: la base del sistema (API Gateway + Auth + User + Postgres) está funcionando y el **billing-service** está en fase inicial (WIP) y todavía **no está integrado al gateway**.

---

## Arquitectura (high level)

- **API Gateway** expone el HTTP público y actúa como reverse proxy hacia los microservicios internos.
- **Auth Service** maneja registro y login (hashing de password, emisión de JWT).
- **User Service** maneja persistencia/consulta de usuarios contra Postgres.
- **PostgreSQL** almacena datos (por ahora la tabla `users`).

Comunicación actual:

- Cliente → API Gateway → Auth Service
- Cliente → API Gateway → User Service
- Auth Service → User Service (para crear usuario y validar credenciales)

El acceso desde el Gateway a los servicios internos se hace mediante un header interno:

- `X-Internal-User-ID`: se usa como “señal” de request interno confiable.

---

## Microservicios actuales

### 1) `api-gateway`
**Responsabilidad:** punto de entrada público. Enruta y proxya requests hacia servicios internos.

- Expone `GET /health`.
- Rutas públicas:
  - `POST /api/auth/register` → `auth-service POST /register`
  - `POST /api/auth/login` → `auth-service POST /login`
- Rutas protegidas (requieren JWT):
  - `GET /api/auth/me` → `auth-service GET /me`
  - `GET/POST /api/users/*` → `user-service /users/*`

**Auth en el gateway:**

- valida JWT (middleware JWT)
- agrega headers internos para llamadas a servicios internos

Archivos clave:
- `services/api-gateway/internal/server/server.go`
- `services/api-gateway/internal/router/router.go`

---

### 2) `auth-service`
**Responsabilidad:** registro y login.

- `POST /register`: genera hash bcrypt y delega creación al `user-service`.
- `POST /login`: consulta usuario por email en `user-service`, compara bcrypt y emite JWT.
- `GET /me`: endpoint protegido por “internal auth” (se espera que lo consuma el gateway).

Archivos clave:
- `services/auth-service/internal/service/auth.go`
- `services/auth-service/internal/client/user_client.go`
- `services/auth-service/internal/server/server.go`

---

### 3) `user-service`
**Responsabilidad:** CRUD de usuarios (por ahora create y get).

- `POST /users`: crea usuario (la password ya llega hasheada desde `auth-service`).
- `GET /users/{id}`: busca usuario por ID.
- `GET /users/email/{email}`: busca usuario por email (incluye password hasheada en la respuesta; se usa para login).

**Seguridad:** sus rutas están protegidas por un middleware interno que exige `X-Internal-User-ID`.

**Persistencia:** PostgreSQL vía `pgxpool`.

Migraciones:
- `services/user-service/migrations/001_create_users.sql`

Archivos clave:
- `services/user-service/internal/db/postgres.go`
- `services/user-service/internal/repository/user_postgres.go`
- `services/user-service/internal/server/server.go`

---

### 4) `billing-service` (WIP)
**Responsabilidad (planeada):** emitir y consultar facturas (invoices), y luego conectar con suscripciones/pagos.

Estado actual:
- Router con endpoints básicos (`POST /invoices`, `GET /invoices`, `GET /health`).
- Conexión a DB hardcodeada (pendiente de config por env/compose).
- No está integrado al `api-gateway` ni al `deploy/docker-compose.yml` todavía.

Archivos clave:
- `services/billing-service/internal/router/router.go`
- `services/billing-service/internal/service/billing.go`

---

## Cómo correr el proyecto (Docker Compose)

La forma recomendada en el estado actual es usar el stack de `deploy/docker-compose.yml`.

1) Crear `deploy/.env` (ver ejemplo en `deploy/README.md`).

2) Levantar servicios:

```bash
cd deploy
docker-compose up -d
```

3) Health checks:

- Gateway: `http://localhost:8080/health`
- User service: `http://localhost:8081/health`
- Auth service: accesible solo internamente en Docker (desde el gateway), pero tiene `GET /health`.

> Nota: en `docker-compose.yml`, `auth-service` no expone puerto hacia el host para forzar el acceso vía gateway.

---

## Endpoints públicos (vía API Gateway)

Base URL: `http://localhost:8080`

### Auth
- `POST /api/auth/register`
  - body: `{ "email": "...", "password": "..." }`
- `POST /api/auth/login`
  - body: `{ "email": "...", "password": "..." }`
  - response: `{ "access_token": "..." }`
- `GET /api/auth/me`
  - header: `Authorization: Bearer <token>`

### Users (protegido)
- `GET /api/users/{id}`
- `GET /api/users/email/{email}`
- `POST /api/users` *(pensado para uso interno; normalmente el alta pública se hace por /api/auth/register)*

---

## Variables de entorno (resumen)

Ver `deploy/README.md` para el detalle.

- Gateway
  - `GATEWAY_HTTP_ADDR` (default `:8080`)
  - `JWT_SECRET`
  - `AUTH_SERVICE_URL` (default `http://auth-service:8082` en compose)
  - `USER_SERVICE_URL` (default `http://user-service:8081`)

- Auth Service
  - `AUTH_HTTP_ADDR` (default `:8080`, en compose se usa `:8082`)
  - `JWT_SECRET`
  - `USER_SERVICE_URL`

- User Service
  - `USER_HTTP_ADDR` (default `:8081`)
  - `USER_DB_DSN` (DSN de Postgres)

---

## Roadmap (alto nivel)

- Integrar `billing-service` al `docker-compose` y al gateway (rutas `/api/billing/*`).
- Modelar multi-tenant (empresas) + suscripciones + planes.
- `payment-service` (integración con MercadoPago / Stripe-like).
- `notification-service` (emails, webhooks y eventos).
- Observabilidad (logs estructurados, tracing, métricas).
- Harden de seguridad (mTLS interno / service auth real / rate limiting / scopes).
