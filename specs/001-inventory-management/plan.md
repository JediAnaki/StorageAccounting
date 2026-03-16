# Implementation Plan: Warehouse Inventory Management via Telegram Mini App

**Branch**: `001-inventory-management` | **Date**: 2026-03-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-inventory-management/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Build a Telegram Mini App for warehouse inventory management that allows warehouse staff to view, add, and update inventory items with photo cards, track quantity changes (additions/reductions), and maintain a complete transaction audit trail. The system prioritizes data integrity with ACID transactions, visual item identification with photo compression, and mobile-first responsive design.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: NEEDS CLARIFICATION - Telegram Bot API library (evaluate go-telegram-bot-api vs alternatives), PostgreSQL driver (pgx vs lib/pq), image processing library for compression/thumbnails, migration tool (golang-migrate)
**Storage**: PostgreSQL (production), SQLite (development/testing), Telegram File API for photos (primary), S3-compatible storage (fallback/archival)
**Testing**: Go testing package with table-driven tests, integration tests against test Telegram bot
**Target Platform**: Linux server (Docker deployment), accessed via Telegram mobile/desktop clients
**Project Type**: Web service (Telegram Bot backend + Mini App frontend served via HTTPS)
**Performance Goals**: <200ms p95 for read operations, <500ms p95 for write operations, <3s for photo upload/compression, support 50 concurrent users
**Constraints**: Mobile-first UI (touch-optimized), HTTPS required for Telegram WebApp, photos <500KB compressed/<50KB thumbnails, ACID transaction compliance
**Scale/Scope**: 50-500 inventory items per warehouse, 10-20 concurrent users peak, 10-100 transactions per item per month, single warehouse deployment initially

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Telegram Mini App Architecture ✅ PASS

**Requirement**: Implement as Telegram Mini App (WebApp) with app-like UX, not CLI bot.

**Compliance**:
- ✅ Project type explicitly: "Web service (Telegram Bot backend + Mini App frontend)"
- ✅ Frontend served via HTTPS (Technical Constraints)
- ✅ Telegram WebApp authentication planned (FR-008 in spec)
- ✅ Mobile-optimized UI required (FR-015 in spec)

**Verdict**: PASS - Architecture fully aligned with Telegram Mini App pattern.

### Principle II: User-Centric Interface Design ✅ PASS

**Requirement**: Visual inventory cards with photos, quantities, transaction history; mobile-first; visual feedback.

**Compliance**:
- ✅ Photo cards for every item (FR-001: "photo thumbnail, item name, current quantity")
- ✅ Transaction history with visual distinction (FR-009: "green for additions (+), red for reductions (-)")
- ✅ Mobile-first design (Constraints: "Mobile-first UI (touch-optimized)")
- ✅ Visual feedback for async operations (FR-012: "loading indicators and progress")
- ⚠️ Offline viewing mentioned in constitution but not in spec - acceptable for MVP, can be progressive enhancement

**Verdict**: PASS - Core UX requirements met; offline support deferred to future iteration.

### Principle III: Data Integrity & Consistency (NON-NEGOTIABLE) ✅ PASS

**Requirement**: ACID transactions, immutable audit log, concurrency control, soft deletes, backups.

**Compliance**:
- ✅ ACID transactions (Storage: "PostgreSQL", Constraints: "ACID transaction compliance")
- ✅ Immutable transaction records (FR-006: "record transaction metadata for every quantity change")
- ✅ Concurrency control (FR-007: "optimistic locking - if conflict detected, reject")
- ✅ Soft deletes (FR-013: "mark as deleted, don't remove from database")
- ⚠️ Automated backups not in functional requirements - MUST be added to implementation tasks

**Verdict**: PASS with action - Core data integrity met; backups must be included in operational tasks.

### Principle IV: Image & Media Handling ✅ PASS

**Requirement**: Photo compression (<500KB), thumbnails (<50KB), Telegram File API primary, S3 fallback, validation.

**Compliance**:
- ✅ Compression targets (FR-003: "compress uploaded photos to under 500KB for full-size display and generate thumbnails under 50KB")
- ✅ Telegram File API + S3 fallback (Storage: "Telegram File API for photos (primary), S3-compatible storage (fallback/archival)")
- ✅ File type validation (FR-011: "validate photo file types (accept only JPEG, PNG, WebP)")
- ✅ User feedback for failures (FR-012: "visual feedback for all async operations")

**Verdict**: PASS - All image handling requirements met.

### Principle V: Idiomatic Go & Simplicity ✅ PASS

**Requirement**: Idiomatic Go, standard library preference, explicit error handling, avoid magic, table-driven tests.

**Compliance**:
- ✅ Language: Go 1.25
- ✅ Testing: "Go testing package with table-driven tests"
- ⚠️ Dependencies need evaluation (NEEDS CLARIFICATION in Technical Context) - must choose minimal, idiomatic libraries
- ✅ Simplicity prioritized (no complex frameworks mentioned)

**Verdict**: PASS - Dependency choices must follow principle during research phase.

### Overall Constitution Compliance: ✅ PASS

**Summary**: All five core principles are satisfied. Two action items:
1. Add automated backup tasks during implementation planning
2. Evaluate dependencies for simplicity during Phase 0 research (already flagged as NEEDS CLARIFICATION)

**No violations requiring justification in Complexity Tracking.**

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point: HTTP server + Telegram bot webhook
├── internal/
│   ├── models/
│   │   ├── item.go                 # Inventory item entity
│   │   ├── transaction.go          # Transaction entity
│   │   └── user.go                 # User identity from Telegram
│   ├── repository/
│   │   ├── item_repo.go            # Item CRUD with optimistic locking
│   │   └── transaction_repo.go     # Transaction append-only log
│   ├── service/
│   │   ├── inventory_service.go    # Business logic: add/update items, record transactions
│   │   ├── photo_service.go        # Image compression, thumbnail generation, Telegram/S3 upload
│   │   └── auth_service.go         # Telegram WebApp initData validation
│   ├── api/
│   │   ├── handlers.go             # HTTP handlers for Mini App API
│   │   └── middleware.go           # Authentication, logging, error handling
│   └── telegram/
│       ├── bot.go                  # Telegram Bot API integration
│       └── webapp.go               # WebApp launch/authentication helpers
├── migrations/
│   ├── 001_create_items.sql
│   ├── 002_create_transactions.sql
│   └── ...
├── go.mod
└── go.sum

frontend/
├── index.html                      # Mini App entry point
├── css/
│   └── styles.css                  # Mobile-first responsive styles
├── js/
│   ├── app.js                      # Main application logic
│   ├── api.js                      # Backend API client
│   ├── ui.js                       # UI components (item cards, forms, modals)
│   └── telegram.js                 # Telegram WebApp SDK integration
└── assets/
    └── icons/                      # UI icons and placeholders

tests/
├── integration/
│   ├── api_test.go                 # End-to-end API tests
│   └── telegram_test.go            # Telegram WebApp flow tests
└── unit/
    ├── service_test.go             # Business logic tests
    └── repository_test.go          # Data access tests

deployments/
├── Dockerfile
├── docker-compose.yml              # Local development setup (backend + PostgreSQL)
└── nginx.conf                      # HTTPS reverse proxy config for frontend
```

**Structure Decision**: Web application architecture with Go backend and vanilla JavaScript frontend.

**Rationale**:
- **Backend (Go)**: Standard Go project layout with `cmd/` for executables, `internal/` for application code (prevents external imports per Go convention), clear separation of concerns (models, repository, service, API layers)
- **Frontend (vanilla JS)**: Lightweight HTML/CSS/JS served as static files - no heavy SPA framework needed for this use case (constitution principle: simplicity). Telegram WebApp SDK integrated directly.
- **Separation**: Backend and frontend clearly separated for independent development/deployment. Backend serves API + handles Telegram bot webhook. Frontend served via HTTPS (nginx) for Telegram WebApp requirements.
- **Testing**: Integration tests for full API flows, unit tests for business logic and data access. Follows Go testing conventions.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
