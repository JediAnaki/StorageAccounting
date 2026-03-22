<!--
SYNC IMPACT REPORT
==================
Version Change: [UNVERSIONED] → 1.0.0
Change Type: INITIAL RATIFICATION
Date: 2026-03-13

Added Sections:
- I. Telegram Mini App Architecture
- II. User-Centric Interface Design
- III. Data Integrity & Consistency
- IV. Image & Media Handling
- V. Idiomatic Go & Simplicity
- Technology Stack
- Bot Development Standards
- Governance

Templates Requiring Updates:
✅ plan-template.md - Constitution Check section will reference these principles
✅ spec-template.md - Requirements will align with UX and data integrity principles
✅ tasks-template.md - Task phases will reflect bot deployment and testing requirements
✅ AGENTS.md - Already reflects Go as active technology

Follow-up TODOs: None - all placeholders resolved
-->

# StorageAccounting Constitution

## Core Principles

### I. Telegram Mini App Architecture

This project MUST be implemented as a Telegram Mini App (WebApp) within a Telegram bot, not as a traditional CLI bot with text commands. The application MUST provide an app-like user experience leveraging Telegram's Mini App API.

**Rationale**: Telegram Mini Apps provide a native app experience within Telegram, allowing rich UI components (cards, forms, galleries) that are essential for warehouse management with visual inventory cards.

**Requirements**:
- Use Telegram Bot API for bot infrastructure
- Implement Telegram WebApp (Mini App) for user interface
- Frontend MUST be served via HTTPS and authenticated through Telegram WebApp
- Backend MUST validate Telegram WebApp authentication tokens
- UI MUST be responsive and mobile-optimized (primary Telegram platform)

### II. User-Centric Interface Design

User interface MUST prioritize clarity, speed, and visual feedback for warehouse operations. Every inventory item MUST have a visual card representation with photo, quantity, and transaction history.

**Rationale**: Warehouse staff need quick visual identification of items and instant access to stock levels. Photos prevent confusion in fast-paced environments.

**Requirements**:
- Every inventory position MUST display: photo (карточка), current quantity, recent additions (поступления), recent reductions (убавления)
- Photo upload/display MUST be optimized for mobile devices
- Quantity changes MUST be visually distinct (color coding: green for additions, red for reductions)
- Transaction history MUST show timestamps and responsible user
- UI MUST support offline viewing of recently accessed items (progressive enhancement)

### III. Data Integrity & Consistency (NON-NEGOTIABLE)

All inventory transactions MUST be atomic, traceable, and immutable. The system MUST maintain a complete audit log of all quantity changes with user attribution and timestamps.

**Rationale**: Manufacturing warehouse accounting requires precise tracking for compliance, cost accounting, and operational decisions. Lost or incorrect inventory data has direct financial and operational consequences.

**Requirements**:
- Database transactions MUST be ACID-compliant
- All quantity changes MUST create immutable transaction records (additions/reductions)
- Transaction records MUST include: user ID, timestamp (UTC), previous quantity, new quantity, change delta, optional note/reason
- Concurrent updates MUST be handled with optimistic locking or similar concurrency control
- Data validation MUST occur at both API and database layers
- Soft deletes MUST be used for inventory positions (never hard delete)
- Backups MUST be automated and tested for restoration

### IV. Image & Media Handling

Photo storage and retrieval MUST be efficient, secure, and integrated with Telegram's infrastructure where possible. Photos MUST be compressed appropriately without sacrificing identification clarity.

**Rationale**: Photos are core to item identification but can consume significant storage and bandwidth if not managed properly.

**Requirements**:
- Photos MUST be compressed to reasonable size (target: <500KB per image) while maintaining clarity
- Thumbnail generation MUST be automatic for list views (target: <50KB)
- Store photos using Telegram File API when feasible (reduces storage costs, leverages CDN)
- Fallback to object storage (S3-compatible) for larger files or archival
- Image uploads MUST validate file type (JPEG, PNG, WebP only) and size limits
- Failed uploads MUST provide clear user feedback and retry mechanism

### V. Idiomatic Go & Simplicity

All Go code MUST follow idiomatic Go conventions and prioritize simplicity over premature abstraction. Use the standard library unless a dependency provides clear, substantial value.

**Rationale**: Go's strength is its simplicity, tooling, and standard library. Complex abstractions (heavy frameworks, excessive interfaces) reduce maintainability and clarity without providing proportional benefit in this domain.

**Requirements**:
- Use `net/http` for HTTP servers unless specific WebSocket/SSE requirements justify a framework
- Use `database/sql` with a driver (e.g., `lib/pq`, `pgx`) directly unless an ORM provides critical query building safety for complex queries
- Error handling MUST use explicit error returns (`if err != nil`) with context wrapping (`fmt.Errorf` with `%w`)
- Avoid "magic" - no reflection-heavy libraries unless absolutely necessary (e.g., JSON marshaling is acceptable)
- Prefer table-driven tests using `testing` package
- Run `go fmt`, `go vet`, and `golangci-lint` before commits
- Keep cyclomatic complexity low - if a function exceeds ~15-20 lines or 3 levels of nesting, consider refactoring

## Technology Stack

**Language**: Go 1.25

**Bot Framework**: Telegram Bot API
- Use official Telegram Bot API HTTP endpoints or a lightweight Go wrapper (e.g., `github.com/go-telegram-bot-api/telegram-bot-api`)
- Telegram WebApp (Mini App) for frontend UI

**Frontend**: HTML/CSS/JavaScript served as Telegram Mini App
- Lightweight framework acceptable (e.g., vanilla JS, Alpine.js, htmx) - avoid heavy SPA frameworks unless justified
- Must integrate Telegram WebApp JS SDK (`telegram-web-app.js`)

**Database**: PostgreSQL (recommended) or SQLite (for development/prototyping)
- ACID transactions required (see Principle III)
- Use migrations tool (e.g., `golang-migrate/migrate`)

**Image Storage**: Telegram File API (primary), S3-compatible storage (fallback/archival)

**Deployment**: Dockerized Go binary, HTTPS reverse proxy (Caddy/Nginx)

**Testing**: Go `testing` package, table-driven tests, integration tests against test Telegram bot

## Bot Development Standards

### Telegram Bot Security
- Bot token MUST be stored in environment variables, never in code or version control
- Webhook endpoint MUST validate Telegram signature/authentication
- WebApp authentication MUST validate `initData` hash according to Telegram documentation
- User permissions MUST be checked for sensitive operations (deletion, bulk changes)

### API Design
- RESTful conventions for backend API consumed by Mini App
- Endpoints: `GET /items`, `GET /items/:id`, `POST /items`, `PUT /items/:id/quantity`, `POST /items/:id/photo`, `GET /items/:id/transactions`
- Use JSON for request/response bodies
- HTTP status codes MUST be semantically correct (200, 201, 400, 401, 404, 500)
- Error responses MUST include human-readable messages suitable for UI display

### Logging & Observability
- Structured logging required (e.g., `log/slog` in Go 1.21+)
- Log levels: DEBUG (development), INFO (production default), WARN, ERROR
- Log all inventory transactions at INFO level (audit trail)
- Log all errors with stack context
- Include request IDs for tracing user actions through logs

### Performance Targets
- API response time: <200ms p95 for read operations, <500ms p95 for write operations
- Photo upload: <3 seconds p95 for compression and storage
- Support 50+ concurrent users without degradation
- Database queries MUST use indexes on frequently queried fields (item ID, user ID, timestamp)

## Governance

This constitution supersedes all other development practices and decisions. Any deviation MUST be documented with justification in the implementation plan's "Complexity Tracking" section.

**Amendment Process**:
1. Proposed changes MUST be documented with rationale
2. Constitution version MUST be incremented semantically:
   - MAJOR: Principle removal or fundamental redefinition
   - MINOR: New principle or section added
   - PATCH: Clarifications, wording improvements, non-semantic changes
3. Amendment MUST update dependent templates (plan, spec, tasks) for consistency
4. Amendment MUST be committed separately with message: `docs: amend constitution to vX.Y.Z (summary)`

**Compliance**:
- All pull requests MUST verify compliance with applicable principles
- Implementation plans MUST include "Constitution Check" section (see plan-template.md)
- Code reviews MUST reference constitution principles when rejecting/requesting changes
- Unjustified complexity or deviation MUST be rejected

**Runtime Guidance**:
- For day-to-day development questions, consult `AGENTS.md` for commands, project structure, and active technologies
- For feature specifications, planning, and task generation, use `/speckit.*` commands
- For AI-assisted development with Claude, follow `.codex/prompts/` workflow

**Version**: 1.0.0 | **Ratified**: 2026-03-13 | **Last Amended**: 2026-03-13
