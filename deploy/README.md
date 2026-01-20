# Docker Compose Setup

Este directorio contiene la configuración de Docker Compose para levantar todos los servicios del proyecto.

## Servicios

- **db**: Base de datos PostgreSQL
- **pgadmin**: Interfaz web para administrar PostgreSQL
- **user-service**: Microservicio de usuarios (puerto 8081)
- **auth-service**: Microservicio de autenticación (puerto 8080)

## Requisitos Previos

1. Docker y Docker Compose instalados
2. Archivo `.env` en el directorio `deploy/` con las siguientes variables:

```env
# Database
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=postgres
DB_HOST=db
DB_PORT=5432

# User Service
USER_HTTP_ADDR=:8081
# Option 1: Provide full DSN
USER_DB_DSN=postgres://postgres:your_password@db:5432/postgres?sslmode=disable
# Option 2: Leave USER_DB_DSN empty and it will be constructed from DB_USER, DB_PASSWORD, DB_NAME, DB_HOST, DB_PORT

# Auth Service
AUTH_HTTP_ADDR=:8080
JWT_SECRET=your-secret-key-change-in-production
USER_SERVICE_URL=http://user-service:8081
```

**Nota sobre USER_DB_DSN**: Puedes proporcionar el DSN completo en `USER_DB_DSN`, o dejarlo vacío y el servicio lo construirá automáticamente usando las variables `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_HOST` y `DB_PORT`.

## Uso

### Levantar todos los servicios

```bash
cd deploy
docker-compose up -d
```

### Ver logs

```bash
# Todos los servicios
docker-compose logs -f

# Servicio específico
docker-compose logs -f auth-service
docker-compose logs -f user-service
```

### Detener servicios

```bash
docker-compose down
```

### Reconstruir imágenes

```bash
docker-compose build
docker-compose up -d
```

### Ejecutar migraciones

Las migraciones se ejecutan automáticamente la primera vez que se crea la base de datos (archivos en `services/user-service/migrations/` se copian a `/docker-entrypoint-initdb.d/`).

Para ejecutar migraciones manualmente después de la primera creación:

```bash
# Conectarse al contenedor de la base de datos
docker-compose exec db psql -U postgres -d postgres

# O ejecutar un script SQL directamente
docker-compose exec -T db psql -U postgres -d postgres < services/user-service/migrations/001_create_users.sql
```

## Endpoints

- **Auth Service**: http://localhost:8080
  - `POST /register` - Registrar usuario
  - `POST /login` - Iniciar sesión
  - `GET /me` - Obtener usuario actual (requiere JWT)
  - `GET /health` - Health check

- **User Service**: http://localhost:8081
  - `POST /users` - Crear usuario
  - `GET /users/{id}` - Obtener usuario por ID
  - `GET /users/email/{email}` - Obtener usuario por email
  - `GET /health` - Health check

- **PgAdmin**: http://localhost:5050
  - Email: admin@example.com
  - Password: admin123

## Notas

- Los servicios se comunican entre sí usando los nombres de servicio de Docker (ej: `user-service:8081`)
- La base de datos expone el puerto 5432 para conexiones externas
- Los healthchecks aseguran que los servicios estén listos antes de que otros servicios dependientes se inicien
