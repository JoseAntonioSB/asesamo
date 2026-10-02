# ASESAMO — Reglas y Estándares de Desarrollo

> Documento vivo. Toda decisión de arquitectura, stack o convención debe estar registrada aquí antes de codificar.
> Última actualización: Octubre 2026.

---

## 1. Stack Tecnológico

### Backend
| Decisión | Elección |
|---|---|
| Lenguaje | **Go (Golang)** |
| Framework HTTP | **chi** (`github.com/go-chi/chi/v5`) |
| Arquitectura | **Clean Architecture** (puertos y adaptadores, casos de uso, repositorios) |
| Base de Datos | **PostgreSQL** en DigitalOcean Managed DB |
| Acceso a Datos | **sqlc** (SQL puro → código Go tipado; NO usar GORM) |
| Migraciones | **golang-migrate** (archivos `.sql` numerados `000001_description.up.sql`) |
| Autenticación | **JWT** stateless (HS256 o RS256) |
| Almacenamiento de archivos | **Cloudflare R2** vía AWS SDK v2 para Go (API S3 compatible) |
| Emails transaccionales | **Resend** (SDK oficial de Go) |
| IA / OCR de documentos | Proveedor de menor costo al momento de implementar (abstracción vía interfaz interna) |

### Frontend
| Decisión | Elección |
|---|---|
| Framework | **Next.js 14+** |
| Router | **App Router** (Server Components por defecto) |
| Estilos | **Tailwind CSS** |
| Componentes UI | **shadcn/ui** (sobre Radix UI) |
| Estado / Async | **TanStack Query** (React Query v5) |
| Estado global simple | Zustand (solo si se necesita, mínimo uso) |

### Infraestructura y DevOps
| Decisión | Elección |
|---|---|
| Hosting Backend | DigitalOcean (Droplet o App Platform) |
| Hosting Frontend | DigitalOcean App Platform o Vercel |
| Storage | Cloudflare R2 |
| CI/CD | **GitHub Actions** |
| Repositorios | **Repos separados**: `asesamo-api` (Go) y `asesamo-web` (Next.js) |
| Entorno local | **Docker Compose** (backend + DB + herramientas auxiliares en contenedores) |
| Monitoreo errores | **Sentry** (Go SDK + Next.js SDK) |
| Logs | **BetterStack** |
| Pagos | **MercadoPago** |

---

## 2. Arquitectura del Backend (Clean Architecture)

### Estructura de carpetas del repo `asesamo-api`

```
asesamo-api/
├── cmd/
│   └── api/
│       └── main.go               # Entry point. Wirings e inyección de dependencias.
├── internal/
│   ├── domain/                   # Capa de dominio (entidades + interfaces puras)
│   │   ├── entity/               # Structs de negocio: Contractor, Plant, AccessLog…
│   │   └── repository/           # Interfaces (contratos) de repositorios
│   ├── usecase/                  # Casos de uso (reglas de negocio)
│   │   ├── contractor/
│   │   ├── access/
│   │   └── document/
│   ├── adapter/
│   │   ├── http/                 # Handlers chi, middlewares, DTOs de request/response
│   │   │   ├── handler/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   ├── postgres/             # Implementaciones de repositorios (sqlc queries)
│   │   └── storage/              # Adaptador Cloudflare R2
│   └── infra/
│       ├── config/               # Configuración vía env vars
│       ├── database/             # Pool de conexiones, transacciones
│       └── logger/               # Wrapper de logger estructurado
├── db/
│   ├── migrations/               # 000001_init.up.sql, 000001_init.down.sql…
│   ├── queries/                  # Archivos .sql para sqlc
│   └── sqlc.yaml                 # Configuración sqlc
├── pkg/                          # Paquetes reutilizables sin dependencias de dominio
│   ├── jwt/
│   ├── validator/
│   └── pagination/
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── Makefile                      # Comandos: make migrate-up, make sqlc, make test…
```

### Flujo de una petición (regla estricta)

```
HTTP Request
    → chi Router
    → Middleware (auth JWT, tenant, logging)
    → Handler (adapter/http)
        → valida DTO de entrada
        → llama al Use Case (interfaz)
    → Use Case (internal/usecase)
        → lógica de negocio pura
        → llama a Repository (interfaz)
    → Repository Implementation (adapter/postgres)
        → ejecuta query sqlc
        → retorna entidad de dominio
    ← respuesta JSON al cliente
```

**Regla:** Los handlers NUNCA hablan directamente a la DB. Los use cases NUNCA importan paquetes HTTP. Los repositorios NUNCA importan use cases.

---

## 3. Multi-Tenancy

- Cada tabla que contenga datos de un cliente (planta) tendrá una columna `tenant_id UUID NOT NULL`.
- Todo query a estas tablas DEBE incluir `WHERE tenant_id = $1` como primera condición.
- El `tenant_id` se extrae del JWT en el middleware de autenticación y se inyecta en el contexto de Go (`context.Context`).
- Un helper centralizado `tenant.FromContext(ctx)` es la única forma de obtener el tenant_id; nunca se pasa como parámetro manual entre funciones.
- **Nunca** se permite un query sin filtro de tenant en tablas multi-tenant. Los tests de integración deben verificar el aislamiento.

---

## 4. Convenciones de Código

### Go (Backend)

- **Naming:** `camelCase` para variables y funciones internas, `PascalCase` para tipos y funciones exportadas. Paquetes en minúscula, una sola palabra.
- **Errores:** Siempre se devuelven como último valor de retorno. Se usa `fmt.Errorf("usecase/contractor: %w", err)` para wrapping con contexto. Nunca `panic` en código de negocio.
- **Comentarios:** En inglés. Todo tipo o función exportada lleva un comentario GoDoc.
- **Interfaces:** Se definen en el paquete que las **consume**, no en el que las implementa (principio de inversión de dependencias).
- **Context:** Toda función que llame a IO (DB, HTTP, storage) recibe `ctx context.Context` como primer argumento.
- **No se usa `init()`** salvo casos excepcionales documentados.
- **Linter:** `golangci-lint` con al menos: `errcheck`, `govet`, `staticcheck`, `revive`, `goimports`.

### TypeScript / Next.js (Frontend)

- **TypeScript estricto:** `strict: true` en `tsconfig.json`. No se permiten `any` explícitos.
- **Naming:** `camelCase` para variables y funciones, `PascalCase` para componentes y tipos/interfaces.
- **Componentes:** Un componente por archivo. El nombre del archivo coincide con el nombre del componente (`ContractorCard.tsx`).
- **Server vs Client Components:** Por defecto todo es Server Component. Solo se añade `"use client"` cuando se necesita estado, efectos o eventos del navegador. Documentar el porqué en un comentario.
- **Fetching:** Toda llamada a la API pasa por una capa de servicios en `lib/api/` que encapsula la URL base y el token JWT. Los componentes nunca hacen `fetch` directamente.
- **Linter:** ESLint con `@typescript-eslint/recommended` + Prettier.

### Idioma del código

- **Todo el código, comentarios y docstrings: en inglés.**
- Los mensajes de error que llegan al usuario final (UI): en **español**.
- Los nombres de entidades del dominio de negocio mexicano se mantienen en español cuando no tienen traducción natural: `Contractor` (no `Contratista`), pero `REPSE`, `SUA`, `DC3`, `NOM` se usan como acrónimos tal cual.

---

## 5. Base de Datos

### Convenciones SQL

- Tablas en `snake_case` plural: `contractors`, `access_logs`, `documents`.
- PKs siempre `UUID` generado con `gen_random_uuid()`, nunca enteros secuenciales en tablas multi-tenant.
- Timestamps: `created_at TIMESTAMPTZ DEFAULT now()` y `updated_at TIMESTAMPTZ DEFAULT now()` en toda tabla. Se actualiza `updated_at` mediante trigger o lógica de repositorio.
- Soft delete: columna `deleted_at TIMESTAMPTZ NULL`. Queries exluyen `WHERE deleted_at IS NULL`.
- FKs siempre declaradas con `ON DELETE RESTRICT` por defecto. Si se necesita cascade, documentar explícitamente.
- Índices: todo `tenant_id`, toda FK y toda columna usada en `WHERE` frecuente debe tener índice. Documentar en la migración el porqué.

### Migraciones (golang-migrate)

- Formato: `{número con 6 dígitos}_{descripción_corta}.up.sql` y `.down.sql`.
- Cada migración es **idempotente** en el `.up.sql` cuando sea posible (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).
- Las migraciones se ejecutan automáticamente en el pipeline de CI/CD al hacer merge a `main`.
- **Nunca** se modifica una migración ya aplicada en producción; siempre se crea una nueva.

---

## 6. API REST

- Versión en la URL: `/api/v1/...`
- Recursos en plural y `kebab-case`: `/api/v1/contractors`, `/api/v1/access-logs`.
- Métodos HTTP semánticos: `GET` (leer), `POST` (crear), `PUT` (reemplazar), `PATCH` (actualizar parcialmente), `DELETE` (eliminar/soft delete).
- Respuestas de éxito: siempre con envelope:
  ```json
  { "data": { ... }, "meta": { "page": 1, "total": 42 } }
  ```
- Respuestas de error: siempre con envelope:
  ```json
  { "error": { "code": "DOCUMENT_EXPIRED", "message": "El documento DC-3 ha vencido." } }
  ```
- Códigos de error internos (`code`) en `UPPER_SNAKE_CASE` en inglés.
- Mensajes de error (`message`) dirigidos al usuario en **español**.
- Paginación: `?page=1&per_page=20` en todos los endpoints de listado. Máximo `per_page=100`.

---

## 7. Autenticación y Autorización (JWT)

- El JWT se firma con secreto almacenado en variable de entorno `JWT_SECRET` (mínimo 32 chars).
- Payload mínimo del token: `{ "sub": "user_uuid", "tenant_id": "tenant_uuid", "role": "admin|manager|guard|contractor", "exp": ... }`.
- Expiración del access token: **1 hora**.
- Refresh token: **7 días**, almacenado en cookie `httpOnly` + `Secure` + `SameSite=Strict`.
- El middleware de chi extrae y valida el JWT en toda ruta protegida e inyecta `user_id`, `tenant_id` y `role` en el contexto.
- Roles del sistema:
  - `super_admin` — acceso total (operador de ASESAMO).
  - `plant_admin` — administrador de una planta cliente.
  - `ehs_manager` — gerente de seguridad (solo lectura y aprobaciones).
  - `guard` — guardia de caseta (solo módulo de acceso).
  - `contractor` — contratista (solo su propio portal).

---

## 8. Storage (Cloudflare R2)

- El cliente de R2 se inicializa usando `aws-sdk-go-v2` apuntando al endpoint de R2.
- Nunca se almacenan paths completos de R2 en la DB; se almacena solo la `key` (path relativo dentro del bucket). La URL completa se genera en tiempo de servicio con `presigned URLs` de duración corta (15 minutos).
- Estructura de keys: `{tenant_id}/{entity_type}/{entity_id}/{filename_uuid}.{ext}`.
  - Ejemplo: `abc-123/contractors/xyz-456/dc3_2026_cert.pdf`
- Tamaño máximo por archivo: **10 MB**. Validar en handler antes de subir.
- Tipos permitidos: `pdf`, `jpg`, `jpeg`, `png`. Rechazar cualquier otro tipo con error `INVALID_FILE_TYPE`.

---

## 9. Testing

### Backend (Go)

- **Unitarios:** Usan interfaces mockeadas (con `github.com/stretchr/testify/mock` o `mockery`). Ningún test unitario hace IO real.
- **Integración:** Usan `testcontainers-go` para levantar un PostgreSQL real en Docker. Cada test de integración corre en su propia transacción y hace rollback al terminar.
- Cobertura mínima aceptable para use cases: **80%**.
- Comando: `make test` corre unitarios. `make test-integration` corre integración (requiere Docker).
- Archivos de test al lado del código: `contractor_usecase_test.go` junto a `contractor_usecase.go`.

### Frontend (Next.js)

- **Unit tests de componentes:** Vitest + React Testing Library.
- **E2E (a agregar en Fase 2):** Playwright.

---

## 10. Variables de Entorno

- Toda configuración sensible va en variables de entorno. **Cero secrets hardcodeados.**
- Archivo `.env.example` en el repo con todas las variables necesarias (sin valores reales).
- Variables de entorno validadas al inicio de la aplicación (`cmd/api/main.go`). Si falta alguna requerida, el proceso termina con error claro.
- Nombres en `UPPER_SNAKE_CASE`.

```bash
# Ejemplo .env.example
APP_ENV=development           # development | staging | production
APP_PORT=8080
DATABASE_URL=postgres://...
JWT_SECRET=
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
RESEND_API_KEY=
SENTRY_DSN=
MERCADOPAGO_ACCESS_TOKEN=
AI_PROVIDER=gemini            # gemini | openai (conmutable sin cambiar código)
AI_API_KEY=
```

---

## 11. Git y Ramas

### Estrategia de ramas

```
main          ← producción (protegida, solo merge vía PR)
staging       ← pre-producción (merge desde develop)
develop       ← integración continua
feature/*     ← nueva funcionalidad (sale de develop)
fix/*         ← corrección de bug (sale de develop o main si es crítico)
chore/*       ← mantenimiento, deps, config
```

### Conventional Commits

Formato: `<tipo>(<scope>): <descripción en inglés, imperativo, minúscula>`

| Tipo | Cuándo usarlo |
|---|---|
| `feat` | Nueva funcionalidad |
| `fix` | Corrección de bug |
| `chore` | Mantenimiento (deps, configs, CI) |
| `docs` | Solo documentación |
| `refactor` | Refactor sin cambio de comportamiento |
| `test` | Agregar o corregir tests |
| `perf` | Mejora de rendimiento |
| `ci` | Cambios en CI/CD |

**Ejemplos:**
```
feat(access): add QR code validation on contractor check-in
fix(document): prevent expired DC3 from bypassing access control
chore(deps): upgrade chi to v5.1.0
test(contractor): add integration test for tenant isolation
```

---

## 12. Makefile (comandos estándar del proyecto)

El repo de backend tendrá un `Makefile` con al menos:

```makefile
make run              # Levanta el servidor local
make build            # Compila el binario
make test             # Tests unitarios
make test-integration # Tests de integración (requiere Docker)
make migrate-up       # Aplica migraciones pendientes
make migrate-down     # Revierte la última migración
make sqlc             # Regenera el código sqlc desde los .sql
make lint             # Corre golangci-lint
make docker-up        # Levanta docker-compose (DB + app)
make docker-down      # Baja docker-compose
```

---

## 13. Seguridad (Reglas No Negociables)

- Toda comunicación en producción via **HTTPS/TLS**. HTTP redirige a HTTPS.
- Inputs del usuario siempre **validados y sanitizados** antes de cualquier query o proceso. Usar `go-playground/validator/v10` en Go.
- **SQL Injection:** Imposible al usar sqlc con queries parametrizadas. Nunca concatenar strings para formar SQL.
- **Rate Limiting:** El router chi debe incluir un middleware de rate limiting (ej. `go-chi/httprate`) en endpoints públicos y de auth.
- **CORS:** Configurado explícitamente; solo dominios de ASESAMO en producción.
- **Headers de seguridad:** `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, `Content-Security-Policy` configurados en middleware.
- **Archivos subidos:** Validar tipo MIME real (no solo extensión) antes de guardar en R2.
- **Logs:** Nunca se loguea información sensible (tokens JWT, contraseñas, números de tarjeta). Los logs de request loguean solo el path, método, status y duración.

---

## 14. Docker Compose (Entorno Local)

El `docker-compose.yml` levantará:

```yaml
services:
  api:         # Go backend (hot-reload con Air)
  web:         # Next.js frontend (npm run dev)
  db:          # PostgreSQL 16
  migrate:     # Contenedor que corre golang-migrate al iniciar
```

Variables de entorno locales cargadas desde `.env` (no commiteado al repo).

---

*Este documento es la fuente de verdad del desarrollo. Cualquier desviación debe discutirse y actualizarse aquí antes de implementarse.*

---

## 15. AI Agent Skills & Workflows

Para garantizar la calidad visual, arquitectónica y de seguridad, cualquier agente de IA trabajando en este repositorio debe utilizar las siguientes **skills locales** (ubicadas en `.agents/skills/`) de acuerdo a la tarea:

### 🎨 Frontend & UI/UX (Next.js + shadcn/ui)
- **`ui-styling`**: Uso obligatorio para construir componentes con **shadcn/ui** y **Tailwind CSS**. Garantiza accesibilidad (Radix UI), layouts responsivos y soporte para Dark Mode.
- **`ui-ux-pro-max` / `high-end-visual-design`**: Para asegurar que el portal web no parezca "genérico" o un template barato. Define jerarquía limpia, tipografía editorial y espaciado correcto.
- **`ask-sonner`**: Usar estrictamente para las notificaciones toast (éxito al subir DC-3, errores de acceso) en Next.js.
- **`emil-design-eng` / `apple-design`**: Para implementar animaciones fluidas (micro-interacciones en botones, modales, transiciones de páginas) que den una sensación *premium* e industrial al dashboard.

### 🛡️ Backend & Seguridad (Go + PostgreSQL)
- **`security-audit`**: Requerido antes de hacer merge de funciones core (ej. autenticación, control de acceso, subida de documentos a R2).
- **OWASP Skills (`broken-access-control`, `injection`, `insecure-design`)**: Aunque algunas skills mencionen otros lenguajes, los principios de validación de tenant (Tenant Isolation), inyección SQL (prevenida por `sqlc`) y fallos lógicos **deben aplicarse estrictamente al código en Go**.

### 🛠️ Flujo de Ingeniería & Refactor
- **`lean-build`**: Para features nuevos (ej. crear el módulo de REPSE). Mantiene el alcance estricto sin sobre-ingeniería.
- **`investigate-first`**: Obligatorio cuando haya un bug (ej. un contratista no puede ver sus documentos, o el JWT falla en producción) antes de empezar a modificar archivos.
- **`surgical-patch`**: Para arreglar bugs muy específicos sin tocar el comportamiento circundante.
- **`safe-refactor`**: Si un *Use Case* de Go crece demasiado, usar esta skill para extraer lógica manteniendo las pruebas intactas.
- **`verify-and-stop`**: Para pruebas de validación final (ej. confirmar que el webhook de MercadoPago actualiza el status correctamente en la BD) sin expandir el alcance de la tarea.
