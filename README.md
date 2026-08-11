# SIMPKL API

SIMPKL API is the backend service for a practical work placement management
platform. It is designed for schools—especially vocational secondary schools
(SMK)—and can also be adopted by other educational institutions that need to
coordinate students, workplace partners, supervisors, placement administration,
documents, readiness checks, reporting, and historical records.

The API is intentionally focused on the administrative lifecycle of practical
work placement (PKL). It does not currently provide attendance tracking, daily
journals, GPS monitoring, check-in/check-out, a student portal, or an employer
operations portal.

## Product capabilities

The backend exposes the following business areas:

- **Authentication and sessions** — login, access-token refresh, logout, and the
  authenticated-user profile.
- **Role-based access control** — users, roles, permissions, role assignment,
  permission assignment, and server-side authorization.
- **Academic master data** — PKL periods, departments/majors, classes, and
  students.
- **Workplace partners** — companies, company contacts, cooperation metadata,
  capacity, and capacity by major.
- **Supervision** — school supervisors, department assignment, status, and
  maximum student capacity.
- **Placement management** — student/company/supervisor relationships, dates,
  division, position, work system, status transitions, notes, and transfers.
- **Administrative readiness** — readiness calculation, completion indicators,
  document-related checks, overrides, and placement progression support.
- **Private document management** — document types, multipart upload, metadata,
  verification, expiry information, secure download, version numbering, and
  superseding previous documents.
- **Document automation** — configurable institution identity, signatories,
  versioned letter templates, collision-safe numbering, completeness previews,
  individual and batch generation, DOCX/PDF letters, XLSX recaps, ZIP packages,
  immutable data snapshots, checksums, and generation history.
- **Reporting** — placement dashboard data and styled JSON, Excel, and PDF
  reports with human-readable status labels.
- **Archiving** — period snapshots and administrative archive records.
- **Auditability** — audit events for important administrative mutations.

## Domain lifecycle

The intended operational flow is:

~~~text
Define period and academic master data
        ↓
Register students, workplace partners, contacts, and supervisors
        ↓
Create and validate student placements
        ↓
Upload and verify required administrative documents
        ↓
Validate source data and generate official letters/recaps
        ↓
Recalculate readiness and resolve outstanding requirements
        ↓
Monitor placement status and generate reports
        ↓
Archive completed periods for historical reference
~~~

Placement transfers preserve history. The previous placement is retained and
marked as transferred, while the new placement records its
previous_placement_id relationship.

## Technology stack

| Concern | Technology |
| --- | --- |
| Language | Go 1.25.5 |
| HTTP framework | Gin |
| ORM and database access | GORM with the MySQL driver |
| Database | MySQL 8.x |
| Authentication | JWT access and refresh tokens |
| Configuration | Viper and environment variables |
| Password security | golang.org/x/crypto password hashing utilities |
| Validation | go-playground/validator and domain validation services |
| Logging | Uber Zap |
| API documentation | OpenAPI YAML and Swagger UI |
| Spreadsheet export | Excelize |
| DOCX generation | Internal standards-compliant OOXML generator |
| PDF export | Internal A4 letter and landscape report generators |
| Batch packaging | ZIP archives stored in private storage |
| Identifiers | UUID |
| Schema management | Versioned SQL migrations |
| Local infrastructure | Docker Compose with MySQL |

## Architecture

SIMPKL API follows a modular-monolith architecture. It keeps domain boundaries
explicit while remaining straightforward to run and deploy as one service.

~~~text
HTTP request
  → request ID, CORS, recovery, logging, and error middleware
  → authentication middleware
  → permission middleware
  → HTTP handler
  → application service
  → repository / GORM query
  → MySQL
~~~

### Responsibility boundaries

- **Handlers** bind requests, select response status codes, and serialize the
  public API response.
- **Services** enforce validation, workflows, status transitions, capacity
  rules, document behavior, readiness calculations, and audit calls.
- **Repositories** encapsulate persistence and reusable CRUD queries.
- **Entities** define persisted domain structures and validation tags.
- **Middleware** handles cross-cutting HTTP concerns, authentication,
  authorization, request IDs, recovery, CORS, and structured logging.
- **Platform packages** provide database connections, JWT services, storage,
  report generation, logging, and other infrastructure adapters.

Standard CRUD resources use the generic repository/service/handler foundation in
internal/shared/crud. Complex workflows such as authentication, placement
transfer, document upload/verification, readiness, reports, and archiving use
dedicated services.

## Repository structure

~~~text
simpkl-api/
├── cmd/
│   ├── api/                  # HTTP server entrypoint
│   └── seed/                 # Super-admin and realistic fixture seeder
├── docs/
│   ├── openapi.yaml          # API contract used by Swagger UI
│   ├── architecture.md
│   ├── api-conventions.md
│   ├── database.md
│   └── permissions.md
├── internal/
│   ├── app/                  # Composition root and route registration
│   ├── config/               # Environment loading and validation
│   ├── middleware/           # HTTP cross-cutting concerns
│   ├── modules/              # Domain modules
│   ├── platform/             # DB, JWT, storage, reports, logging
│   └── shared/               # CRUD, response, pagination, errors, types
├── migrations/               # Up/down SQL schema migrations
├── scripts/                  # Development and maintenance helpers
├── storage/private/          # Local private document storage
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
~~~

### Domain modules

The internal/modules directory contains auth, periods, majors, classes, students,
companies, companycontacts, supervisors, placements, documents, readiness,
reports, archives, users, roles, permissions, and auditlogs.

Each module normally contains delivery/HTTP, entity, service, repository, and
validation code where those layers are needed.

## API conventions

The versioned API prefix is /api/v1.

Successful responses follow this shape:

~~~json
{
  "success": true,
  "message": "Request completed",
  "data": {},
  "meta": null
}
~~~

Error responses include a stable error code and request ID:

~~~json
{
  "success": false,
  "message": "The submitted data is invalid",
  "code": "VALIDATION_ERROR",
  "errors": {},
  "request_id": "..."
}
~~~

List endpoints support server-side pagination using page, per_page, and
resource-specific filters. per_page is capped at 100. Search parameters are
only applied to explicitly configured columns.

Common HTTP statuses are 200, 201, 401, 403, 404, 409, 422, and 500.
Business conflicts such as invalid placement transitions or exceeded
company/supervisor capacity return 409 with a domain-specific code.

## Endpoint map

All paths below are relative to /api/v1 and protected paths require a bearer
access token unless stated otherwise.

| Area | Main operations |
| --- | --- |
| Auth | POST /auth/login, /auth/refresh, /auth/logout, GET /auth/me |
| Periods | CRUD /periods |
| Majors | CRUD /majors |
| Classes | CRUD /classes |
| Students | CRUD /students, POST /students/import |
| Companies | CRUD /companies, major-capacity operations |
| Company contacts | CRUD /company-contacts |
| Supervisors | CRUD /supervisors |
| Placements | CRUD /placements, POST /placements/{id}/transfer |
| Documents | CRUD metadata, upload, verification, secure download, document types |
| Document automation | Profile/signatory/template management, preview, generate, history, DOCX/PDF/XLSX/ZIP download |
| Readiness | list, recalculate, override |
| Reports | dashboard and placement reports in JSON/XLSX/PDF |
| Archives | list, create, and detail |
| RBAC | CRUD /users, /roles, and /permissions |
| Audit | protected audit-log listing |

Additional request and response details are maintained in
[docs/openapi.yaml](docs/openapi.yaml).

## Security model

- Access tokens are sent through Authorization: Bearer <token>.
- Refresh tokens are rotated when refreshed and can be revoked on logout.
- Backend permission checks are authoritative; hiding a frontend navigation item
  is not treated as authorization.
- RBAC permissions are grouped by domain, such as student.view,
  placement.update, document.verify, and report.view.
- Private documents are stored outside public static assets and are served only
  through an authenticated, permission-checked download endpoint.
- Error responses avoid exposing internal database or filesystem details.
- Change-sensitive administrative operations can carry X-Change-Reason.
- Secrets must be supplied through environment configuration and must never be
  committed.

## Database and storage

Production schema changes are managed through SQL migrations, not AutoMigrate.

Important relationships include:

~~~text
periods → placements ← students
companies → company_contacts
companies → placements ← supervisors
placements → documents
placements → generated_documents ← document_templates
document_generation_batches → generated_documents
school_profiles + signatories → generated letter identity
students + periods → administrative_readiness
periods → archives
users → user_roles → roles → role_permissions → permissions
users → refresh_sessions and audit_logs
~~~

Soft deletion is used for applicable operational and master data. Optional
foreign-key values are normalized to SQL NULL when a relationship is not
selected, preserving ON DELETE SET NULL behavior.

## Local development

### Prerequisites

- Go 1.25.5 or a compatible Go toolchain.
- MySQL 8.x, or Docker Desktop for the included Compose setup.
- The migration CLI available through the Go toolchain.

### Setup with Docker Compose

~~~powershell
docker compose up -d mysql
Copy-Item .env.example .env
~~~

Update .env with development values. Do not reuse these placeholders for a
shared or production environment.

### Apply migrations

By default, the API automatically applies embedded versioned migrations before
it starts listening. The SQL files are compiled into the API binary, and the
runner uses `schema_migrations` plus a MySQL advisory lock to prevent concurrent
application by multiple API instances.

```text
MIGRATIONS_AUTO_APPLY=true
```

For deployments that use a separate migration job, set
`MIGRATIONS_AUTO_APPLY=false` on the API and run the existing migration CLI.
The project uses a MySQL URL compatible with golang-migrate:

~~~text
mysql://user:password@tcp(localhost:3306)/simpkl?multiStatements=true
~~~

Then run the separate migration job:

~~~powershell
go run github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database "$env:DATABASE_URL" up
~~~

### Seed local data

The seeder creates a realistic, idempotent PKL dataset for development. It
includes periods, majors, classes, students, workplace partners, contacts,
supervisors, placements, uploaded documents, automation templates, an
institution profile, a default signatory, readiness records, archives, RBAC
data, and an administrative user. Record counts are context-aware rather than
forcing the same number into every table.

Configure seed behavior in .env:

~~~text
SEED_ENABLED=true
SEED_RECORD_COUNT=5
SEED_RESET_LEGACY=true
SEED_ADMIN_NAME=Super Admin SIMPKL
SEED_ADMIN_EMAIL=admin@example.sch.id
SEED_ADMIN_USERNAME=superadmin
SEED_ADMIN_PASSWORD=replace_with_at_least_8_characters
~~~

Run:

~~~powershell
go run ./cmd/seed
~~~

The fixture data is for local development and demonstration only. It must not
be treated as real student, school, or company data.

### Run the API

~~~powershell
go run ./cmd/api
~~~

Useful URLs:

- Health: http://localhost:8080/health
- Versioned health: http://localhost:8080/api/v1/health
- OpenAPI file: http://localhost:8080/openapi.yaml
- Swagger UI: http://localhost:8080/swagger/index.html

## Document automation workflow

The automation engine is deterministic: it only merges validated database data
into an active template. It does not invent names, dates, addresses, positions,
or official decisions.

1. Complete the institution profile and select a default active signatory.
2. Review the seeded templates or create a new template/version using supported
   placeholders.
3. Select a period and optionally narrow by class, major, company, supervisor,
   or a single placement.
4. Preview completeness. Generation is blocked when required official data is
   missing.
5. Generate DOCX/PDF letters and an XLSX recap. The API stores every file in
   private storage and creates one ZIP package per batch.
6. Download the package or individual files from generation history.

Each generated record stores the template code/version, official number,
requesting user, data snapshot, SHA-256 checksum, storage metadata, and creation
time. Editing master data later therefore does not rewrite historical output.
Letter sequences are scoped by template, year, and month and incremented in a
database transaction.

The default templates provide an introduction/application letter, placement
letter, supervisor assignment letter, parent/guardian consent letter, and
placement recap workbook. Template content and institution identity remain
configurable so the product is not tied to a particular school.

See [docs/document-automation.md](docs/document-automation.md) for placeholders,
security behavior, and the generation lifecycle.

## Environment variables

The complete safe template is in [.env.example](.env.example). Main groups are:

- APP_* — service name, environment, port, and public URL.
- DB_* — MySQL connection and pool settings.
- JWT_* — access/refresh secrets and token lifetimes.
- CORS_ALLOWED_ORIGINS — browser origins allowed to call the API.
- STORAGE_* — private file-storage driver and path.
- LOG_LEVEL — structured logging level.
- SEED_* — local fixture and super-admin bootstrap behavior.

Never expose JWT secrets, database passwords, or storage credentials in frontend
variables or source control.

## Commands

~~~text
make run          Run the API
make build        Build the API binary
make test         Run all Go tests
make vet          Run go vet
make fmt          Format Go source
make migrate-up   Apply migrations using DATABASE_URL
make migrate-down Roll back one migration
make seed         Run the local fixture seeder
make docker-up    Start the local MySQL container
make docker-down  Stop the local containers
~~~

## Validation

Before opening a change for review, run the checks relevant to the area:

~~~powershell
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
go build -o ./bin/simpkl-api.exe ./cmd/api
git diff --check
~~~

For schema or API contract changes, also review the migration, OpenAPI file,
permission behavior, error shape, and affected frontend consumers.

## Documentation map

- [docs/architecture.md](docs/architecture.md) — modular-monolith boundaries
  and request flow.
- [docs/api-conventions.md](docs/api-conventions.md) — response, pagination,
  authentication, and status conventions.
- [docs/database.md](docs/database.md) — schema relationships and lifecycle.
- [docs/permissions.md](docs/permissions.md) — RBAC permission groups.
- [docs/openapi.yaml](docs/openapi.yaml) — machine-readable API contract.

## Scope and roadmap boundary

The current service is an administrative foundation for PKL coordination. A
future release may add attendance, journals, notifications, student access,
employer access, or richer dashboard analytics, but those capabilities are not
part of the current API contract and should not be inferred from existing
placement or readiness endpoints.

## License

See [LICENSE](LICENSE).
