# SaaS Subscription Platform (mini Stripe/Paddle)

Plataforma SaaS orientada a empresas que cobran **suscripciones**. El objetivo del repo es construir un conjunto de microservicios (estilo Stripe/Paddle/MercadoPago para SaaS) para:

- registrar usuarios
- autenticar y emitir JWT
- gestionar suscripciones y facturación
- (a futuro) pagos, notificaciones, webhooks, etc.

> Estado actual: la base del sistema (API Gateway + Auth + User + Postgres) está funcionando y el **billing-service** ya está integrado detrás del API Gateway (rutas `/api/billing/*`).

---

## Arquitectura (high level)

- **API Gateway** expone el HTTP público y actúa como reverse proxy hacia los microservicios internos.
- **Auth Service** maneja registro y login (hashing de password, emisión de JWT).
- **User Service** maneja persistencia/consulta de usuarios contra Postgres.
- **Billing Service** maneja facturas (invoices): creación y listado (MVP).
- **PostgreSQL** almacena datos (por ahora `users` e `invoices`).

Comunicación actual:

- Cliente → API Gateway → Auth Service
- Cliente → API Gateway → User Service
- Cliente → API Gateway → Billing Service
- Auth Service → User Service (para crear usuario y validar credenciales)
- Billing Service → Postgres (tabla `invoices`)

### Headers internos + trazabilidad

El accesso desde el Gateway a los servicios internos se hace mediante headers internos.

- `X-Internal-User-ID`: señal de request interno confiable (los servicios internos la exigen).
- `X-Internal-Request-ID`: ID de request para correlación end-to-end (generado/propagado por el gateway).
- `X-Internal-Call-Stack`: “stack”/cadena de hops del request para debugging (ej: `api-gateway>auth-service>user-service`).

Código relacionado:
- Helper/contrato de trazabilidad: `libs/trace/trace.go`
- Inyección de headers internos en gateway: `services/api-gateway/internal/middleware/internal_headers.go`

---

## Microservicios actuales

### 1) `api-gateway`
**Responsabilidad:** punto de entrada público. Enruta y proxea requests hacia servicios internos.

- Expone `GET /health`.
- Rutas públicas:
  - `POST /api/auth/register` → `auth-service POST /register`
  - `POST /api/auth/login` → `auth-service POST /login`
- Rutas protegidas (requieren JWT):
  - `GET /api/auth/me` → `auth-service GET /me`
  - `GET/POST /api/users/*` → `user-service /users/*`
  - `GET/POST /api/billing/*` → `billing-service /*`

**Auth en el gateway:**

- valida JWT (middleware JWT)
- agrega headers internos para llamadas a servicios internos (`X-Internal-User-ID`, `X-Internal-Request-ID`, `X-Internal-Call-Stack`)

Archivos clave:
- `services/api-gateway/internal/server/server.go`
- `services/api-gateway/internal/router/router.go`

---

### 2) `auth-service`
**Responsabilidad:** registro y login.

- `POST /register`: genera hash bcrypt y delega creación al `user-service`.
- `POST /login`: consulta usuario por email en `user-service`, compara bcrypt y emite JWT.
- `GET /me`: endpoint protegido por “internal auth” (se espera que lo consuma el gateway).

Notas de trazabilidad:
- El middleware interno lee `X-Internal-Request-ID` / `X-Internal-Call-Stack` y agrega `auth-service` al stack.
- Las llamadas auth → user propagan trazabilidad vía `context`.

---

### 3) `user-service`
**Responsabilidad:** CRUD de usuarios (por ahora create y get).

- `POST /users`: crea usuario (la password ya llega hasheada desde `auth-service`).
- `GET /users/{id}`: busca usuario por ID.
- `GET /users/email/{email}`: busca usuario por email (incluye password hasheada en la respuesta; se usa para login).

**Seguridad:** protegido por middleware interno que exige `X-Internal-User-ID`.

Notas de trazabilidad:
- El middleware interno lee `X-Internal-Request-ID` / `X-Internal-Call-Stack` y agrega `user-service` al stack.

---

### 4) `billing-service` (integrado)
**Responsabilidad:** facturación (MVP). Crea y lista facturas (invoices) persistidas en Postgres.

Endpoints internos del servicio:
- `GET /health`
- `POST /invoices` *(requiere header interno)*
- `GET /invoices` *(requiere header interno)*

**Cómo funciona (MVP):**
- El cliente llama al gateway en `/api/billing/...` con JWT.
- El gateway valida el JWT y agrega `X-Internal-User-ID`.
- El billing-service valida el header interno y ejecuta la operación contra Postgres.

Persistencia / migraciones:
- Migración: `services/billing-service/migrations/001_create_invoices.sql`
- Tabla: `invoices`

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
- Billing service: accesible solo internamente desde el gateway, pero tiene `GET /health`.

---

## Endpoints públicos (vía API Gateway)

Base URL: `http://localhost:8080`

### Auth
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/me` (JWT)

### Users (protegido)
- `GET /api/users/{id}`
- `GET /api/users/email/{email}`

### Billing (protegido)
- `POST /api/billing/invoices`
  - body: `{ "user_id": "<uuid>", "amount": 123.45 }`
- `GET /api/billing/invoices`

---

## Observabilidad / Debugging (logs)

El sistema implementa trazabilidad “liviana” (sin tracing distribuido completo) basada en `request_id` y `call_stack`.

### Eventos de log principales

- `request_start` / `request_end` (auth-service y user-service)
  - Campos: `service`, `method`, `path`, `status` (en end), `duration_ms`, `request_id`, `call_stack`.
- `upstream_call` / `upstream_call failed` (auth-service → user-service)
  - Campos: `service=user-service`, `method`, `path`, `status` (si hay response), `duration_ms`, `request_id`, `call_stack`, `err`.

### Cómo debuggear un request end-to-end

1) Tomá un `request_id` de cualquier log (ej: `request_end ... request_id=...`).
2) Filtrá logs por ese `request_id` en todos los servicios.
3) Usá `call_stack` para entender por qué servicios pasó el request (hops) y dónde falló.

Código relacionado:
- Middleware de logging: `services/auth-service/internal/middleware/request_logger.go`, `services/user-service/internal/middleware/request_logger.go`
- Cliente auth → user: `services/auth-service/internal/client/user_client.go`

---

## Variables de entorno (resumen)

- Gateway
  - `GATEWAY_HTTP_ADDR` (default `:8080`)
  - `JWT_SECRET`
  - `AUTH_SERVICE_URL`
  - `USER_SERVICE_URL`
  - `BILLING_SERVICE_URL`

- Billing Service
  - `BILLING_HTTP_ADDR` (default `:8083`)
  - `BILLING_DB_DSN`

---

## Roadmap (alto nivel)

- Endpoints de invoices más completos (`GET /invoices/{id}`, filtros, paginado).
- Mejorar modelo de dinero (evitar `float64`, usar centavos + moneda).
- Modelar multi-tenant (empresas) + suscripciones + planes.
- `payment-service` (integración con MercadoPago / Stripe-like).
- `notification-service` (emails, webhooks y eventos).
- Observabilidad (logs estructurados, tracing, métricas).
- Harden de seguridad (mTLS interno / service auth real / rate limiting / scopes).
