# Tasks: Warehouse Inventory Management via Telegram Mini App

**Input**: Design documents from `/specs/001-inventory-management/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Tests**: Tests are not explicitly requested in the specification, so test tasks are omitted. Focus is on implementation tasks.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

Based on plan.md, this project uses:
- **Backend**: `backend/` (Go service)
- **Frontend**: `frontend/` (vanilla JS Mini App)
- **Tests**: `tests/` (integration and unit tests)
- **Deployments**: `deployments/` (Docker, nginx config)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project directory structure per plan.md (backend/, frontend/, tests/, deployments/, migrations/)
- [X] T002 Initialize Go module in backend/go.mod with Go 1.25
- [X] T003 [P] Create backend entry point in backend/cmd/server/main.go with basic HTTP server
- [X] T004 [P] Create frontend entry point in frontend/index.html with Telegram WebApp SDK integration
- [X] T005 [P] Setup PostgreSQL database schema migrations framework in backend/migrations/
- [X] T006 [P] Create Docker Compose configuration in deployments/docker-compose.yml for local development (backend + PostgreSQL)
- [X] T007 [P] Create nginx configuration in deployments/nginx.conf for HTTPS reverse proxy
- [X] T008 [P] Create Dockerfile in deployments/Dockerfile for backend deployment

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T009 Create database migration 001_create_items.sql in backend/migrations/ with items table (id, name, current_quantity, photo_file_id, photo_s3_key, created_at, updated_at, deleted_at, created_by, version for optimistic locking)
- [X] T010 [P] Create database migration 002_create_transactions.sql in backend/migrations/ with transactions table (id, item_id, user_id, timestamp, previous_quantity, new_quantity, delta, note)
- [X] T011 [P] Implement base models: User in backend/internal/models/user.go (Telegram user ID, display name)
- [X] T012 [P] Implement Item model in backend/internal/models/item.go (all attributes from schema)
- [X] T013 [P] Implement Transaction model in backend/internal/models/transaction.go (all attributes from schema)
- [X] T014 Setup PostgreSQL connection pool in backend/internal/repository/db.go with pgx driver
- [X] T015 [P] Implement Telegram WebApp authentication middleware in backend/internal/api/middleware.go (validate initData per Telegram spec)
- [X] T016 [P] Implement error handling middleware in backend/internal/api/middleware.go (structured JSON error responses)
- [X] T017 [P] Implement logging middleware in backend/internal/api/middleware.go using standard library logger
- [X] T018 [P] Setup API routing structure in backend/internal/api/handlers.go with base handler registration
- [X] T019 [P] Implement environment configuration loader in backend/cmd/server/main.go (database URL, Telegram bot token, S3 credentials, server port)
- [X] T020 [P] Create Telegram bot initialization in backend/internal/telegram/bot.go (webhook setup, basic bot commands)
- [X] T021 [P] Implement Telegram WebApp launch helper in backend/internal/telegram/webapp.go (generate Mini App URL)
- [X] T022 [P] Create frontend API client base in frontend/js/api.js (fetch wrapper with authentication headers)
- [X] T023 [P] Create frontend Telegram SDK integration in frontend/js/telegram.js (WebApp.initData, user info, theme)
- [X] T024 [P] Create base UI component framework in frontend/js/ui.js (modal, loading indicator, toast notification helpers)
- [X] T025 [P] Create mobile-first CSS framework in frontend/css/styles.css (touch-optimized buttons, responsive grid, color scheme)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - View Inventory Items (Priority: P1) 🎯 MVP

**Goal**: Warehouse staff can browse all inventory items with photos, names, and current quantities

**Independent Test**: Add sample inventory items to database, open Telegram Mini App, verify scrollable list with all items showing photos, names, and quantities. Verify out-of-stock items are visually distinct. Verify tapping an item shows transaction history.

### Implementation for User Story 1

- [X] T026 [P] [US1] Implement ItemRepository.ListAll() in backend/internal/repository/item_repo.go with pagination support (limit, offset)
- [X] T027 [P] [US1] Implement ItemRepository.GetByID() in backend/internal/repository/item_repo.go with transaction history join
- [X] T028 [P] [US1] Implement TransactionRepository.ListByItemID() in backend/internal/repository/transaction_repo.go (reverse chronological, pagination)
- [X] T029 [US1] Implement InventoryService.ListItems() in backend/internal/service/inventory_service.go (calls ItemRepository, handles pagination)
- [X] T030 [US1] Implement InventoryService.GetItemDetails() in backend/internal/service/inventory_service.go (item + transactions, pagination)
- [X] T031 [US1] Implement GET /api/items endpoint in backend/internal/api/handlers.go (auth required, pagination params, returns JSON list)
- [X] T032 [US1] Implement GET /api/items/:id endpoint in backend/internal/api/handlers.go (auth required, returns item + transactions)
- [X] T033 [US1] Implement inventory list rendering in frontend/js/app.js (fetch items, render photo cards with name/quantity)
- [X] T034 [US1] Implement item card component in frontend/js/ui.js (photo thumbnail, name, quantity, visual distinction for zero stock)
- [X] T035 [US1] Implement infinite scroll pagination in frontend/js/app.js (lazy load 50 items at a time)
- [X] T036 [US1] Implement item detail modal in frontend/js/ui.js (full photo, name, quantity, transaction history list)
- [X] T037 [US1] Style inventory cards in frontend/css/styles.css (touch-friendly, responsive grid, grayed-out for zero stock)
- [X] T038 [US1] Implement transaction history rendering in frontend/js/app.js (green for additions, red for reductions, timestamp, user, note)
- [X] T039 [US1] Implement transaction history pagination in frontend/js/app.js (infinite scroll for history)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently - users can view inventory list, see item details, and review transaction history

---

## Phase 4: User Story 2 - Add New Inventory Item (Priority: P2)

**Goal**: Warehouse managers can add new items with name, photo, and initial quantity

**Independent Test**: Open Mini App, click "Add Item" button, fill in name/photo/quantity, submit form, verify item appears in inventory list immediately with compressed photo

### Implementation for User Story 2

- [X] T040 [P] [US2] Implement photo validation in backend/internal/service/photo_service.go (file type check: JPEG/PNG/WebP only, max 10MB)
- [X] T041 [P] [US2] Implement photo compression in backend/internal/service/photo_service.go (resize to <500KB full-size, generate <50KB thumbnail using image processing library)
- [X] T042 [P] [US2] Implement Telegram File API upload in backend/internal/service/photo_service.go (upload compressed photo, get file_id)
- [X] T043 [P] [US2] Implement S3 fallback upload in backend/internal/service/photo_service.go (upload to S3-compatible storage, get object key)
- [X] T044 [US2] Implement ItemRepository.Create() in backend/internal/repository/item_repo.go (insert item with photo references, return created item with ID)
- [X] T045 [US2] Implement TransactionRepository.Create() in backend/internal/repository/transaction_repo.go (insert transaction record, immutable append-only)
- [X] T046 [US2] Implement InventoryService.AddItem() in backend/internal/service/inventory_service.go (validate inputs, compress photo, create item, create initial transaction record in ACID transaction)
- [X] T047 [US2] Implement POST /api/items endpoint in backend/internal/api/handlers.go (auth required, multipart form with name/photo/quantity, returns created item JSON)
- [X] T048 [US2] Implement "Add Item" button in frontend/js/app.js (opens add item modal)
- [X] T049 [US2] Implement add item form in frontend/js/ui.js (name input, photo file picker, quantity input, submit button)
- [X] T050 [US2] Implement photo upload with progress in frontend/js/app.js (multipart form submission, progress indicator, error handling)
- [X] T051 [US2] Implement form validation in frontend/js/app.js (required fields check, clear error messages for missing photo/name)
- [X] T052 [US2] Style add item form in frontend/css/styles.css (mobile-friendly inputs, touch-optimized file picker)
- [X] T053 [US2] Implement optimistic UI update in frontend/js/app.js (immediately add item to list on submit, show loading state during upload)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently - users can view inventory and add new items with photos

---

## Phase 5: User Story 3 - Record Quantity Changes (Priority: P3)

**Goal**: Warehouse staff can record additions and reductions with quantity changes and optional notes

**Independent Test**: Select existing item, choose "Add Stock" and enter +50, verify quantity updates and green transaction appears. Choose "Remove Stock" and enter -30, verify quantity updates and red transaction appears. Test concurrent updates show conflict error.

### Implementation for User Story 3

- [X] T054 [P] [US3] Implement ItemRepository.UpdateQuantityWithLock() in backend/internal/repository/item_repo.go (optimistic locking using version field, UPDATE WHERE id = ? AND version = ?)
- [X] T055 [US3] Implement InventoryService.RecordAddition() in backend/internal/service/inventory_service.go (validate positive delta, update item quantity with lock, create transaction record in ACID transaction)
- [X] T056 [US3] Implement InventoryService.RecordReduction() in backend/internal/service/inventory_service.go (validate delta, check for negative quantity warning, update item with lock, create transaction in ACID transaction)
- [X] T057 [US3] Implement PUT /api/items/:id/add endpoint in backend/internal/api/handlers.go (auth required, JSON body with quantity delta and optional note)
- [X] T058 [US3] Implement PUT /api/items/:id/remove endpoint in backend/internal/api/handlers.go (auth required, JSON body with quantity delta and optional note)
- [X] T059 [US3] Implement optimistic locking error handling in backend/internal/api/handlers.go (detect version mismatch, return 409 Conflict with user-friendly message)
- [X] T060 [US3] Implement "Add Stock" button in frontend/js/ui.js (item detail modal, opens quantity change form)
- [X] T061 [US3] Implement "Remove Stock" button in frontend/js/ui.js (item detail modal, opens quantity change form)
- [X] T062 [US3] Implement quantity change form in frontend/js/ui.js (quantity input, optional note textarea, submit button)
- [X] T063 [US3] Implement negative quantity warning in frontend/js/app.js (show warning when reduction exceeds current quantity, allow user to proceed or cancel)
- [X] T064 [US3] Implement concurrent update error handling in frontend/js/app.js (detect 409 Conflict response, show "Item was updated by another user, please refresh" message)
- [X] T065 [US3] Implement optimistic UI update in frontend/js/app.js (immediately update quantity on submit, show loading state, rollback on error)
- [X] T066 [US3] Style quantity change form in frontend/css/styles.css (clear visual distinction for add vs remove, touch-friendly inputs)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently - users can view inventory, add items, and track quantity changes with conflict detection

---

## Phase 6: User Story 4 - View Transaction History (Priority: P4)

**Goal**: View all historical quantity changes with timestamps, users, and notes in chronological order

**Independent Test**: View item detail page for an item with 20+ transactions, verify all transactions are listed newest-first with timestamp/user/delta/note, verify green for additions and red for reductions, verify pagination loads older transactions

### Implementation for User Story 4

**Note**: Transaction history display is already implemented in US1 (T038, T039). This phase adds enhancements for better audit trail visibility.

- [X] T067 [P] [US4] Implement transaction filtering in backend/internal/repository/transaction_repo.go (filter by date range, transaction type: addition/reduction)
- [X] T068 [P] [US4] Implement InventoryService.GetTransactionHistory() in backend/internal/service/inventory_service.go (supports filtering, pagination, sorting)
- [X] T069 [US4] Extend GET /api/items/:id/transactions endpoint in backend/internal/api/handlers.go (add filter params: start_date, end_date, type)
- [ ] T070 [US4] Implement transaction filter UI in frontend/js/ui.js (date range picker, type filter: all/additions/reductions)
- [ ] T071 [US4] Implement transaction export in frontend/js/app.js (download transaction history as CSV for item)
- [ ] T072 [US4] Style transaction history in frontend/css/styles.css (clear timestamp formatting, user attribution, visual delta indicators)

**Checkpoint**: All user stories should now be independently functional - complete inventory management with full audit trail

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and operational requirements

- [ ] T073 [P] Implement soft delete for items in backend/internal/service/inventory_service.go (set deleted_at timestamp, preserve in database)
- [ ] T074 [P] Implement DELETE /api/items/:id endpoint in backend/internal/api/handlers.go (soft delete, auth required)
- [ ] T075 [P] Filter deleted items from inventory list in backend/internal/repository/item_repo.go (WHERE deleted_at IS NULL)
- [ ] T076 [P] Implement automated database backups in deployments/backup.sh (PostgreSQL pg_dump scheduled via cron)
- [ ] T077 [P] Add health check endpoint GET /health in backend/internal/api/handlers.go (database connectivity, Telegram API connectivity)
- [ ] T078 [P] Add metrics endpoint GET /metrics in backend/internal/api/handlers.go (active users, items count, transactions count)
- [ ] T079 [P] Implement rate limiting middleware in backend/internal/api/middleware.go (per-user request limits)
- [ ] T080 [P] Add photo storage quota monitoring in backend/internal/service/photo_service.go (track total storage used)
- [ ] T081 [P] Implement error recovery for photo upload failures in frontend/js/app.js (retry button, queue for offline uploads)
- [ ] T082 [P] Add loading skeletons for inventory list in frontend/css/styles.css (improve perceived performance)
- [ ] T083 [P] Implement search/filter for inventory list in frontend/js/app.js (filter by name, filter by stock status: in-stock/out-of-stock)
- [ ] T084 [P] Add favicon and app icons in frontend/assets/icons/ (Telegram WebApp icon requirements)
- [ ] T085 [P] Create README.md in repository root with setup instructions, architecture overview, deployment guide
- [ ] T086 Run migration and verify database schema is correct
- [ ] T087 Deploy to test environment and validate Telegram WebApp authentication flow
- [ ] T088 Load test with 50 concurrent users and verify performance targets (<200ms p95 reads, <500ms p95 writes)
- [ ] T089 Security audit: validate all endpoints require authentication, check for SQL injection vulnerabilities, validate HTTPS enforcement
- [ ] T090 Code cleanup: remove unused imports, add inline documentation, format with gofmt

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-6)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3 → P4)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent of US1 but integrates with inventory display
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Depends on Item model from US1/US2 but should be independently testable
- **User Story 4 (P4)**: Extends US1 transaction history - Can start after Foundational (Phase 2)

### Within Each User Story

- Repository layer before service layer
- Service layer before API handlers
- API handlers before frontend integration
- Core implementation before UI polish
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel (T003-T008)
- All Foundational tasks marked [P] can run in parallel within their layer:
  - Migrations (T009, T010) in parallel
  - Models (T011, T012, T013) in parallel
  - Middleware (T015, T016, T017) in parallel
  - Frontend base (T022, T023, T024, T025) in parallel
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- Within US1: Repository methods (T026, T027, T028) in parallel
- Within US2: Photo service methods (T040, T041, T042, T043) in parallel
- Within US3: Repository and service methods (T054, T055, T056) can be developed in parallel
- All Polish tasks marked [P] can run in parallel (T073-T085)

---

## Parallel Example: User Story 1

```bash
# Launch all repository methods for User Story 1 together:
Task: "Implement ItemRepository.ListAll() in backend/internal/repository/item_repo.go"
Task: "Implement ItemRepository.GetByID() in backend/internal/repository/item_repo.go"
Task: "Implement TransactionRepository.ListByItemID() in backend/internal/repository/transaction_repo.go"

# After repository layer is done, service layer:
Task: "Implement InventoryService.ListItems() in backend/internal/service/inventory_service.go"
Task: "Implement InventoryService.GetItemDetails() in backend/internal/service/inventory_service.go"
```

---

## Parallel Example: User Story 2

```bash
# Launch all photo service methods together:
Task: "Implement photo validation in backend/internal/service/photo_service.go"
Task: "Implement photo compression in backend/internal/service/photo_service.go"
Task: "Implement Telegram File API upload in backend/internal/service/photo_service.go"
Task: "Implement S3 fallback upload in backend/internal/service/photo_service.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (View Inventory)
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Deploy/demo if ready - Users can browse inventory items

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP: View-only inventory)
3. Add User Story 2 → Test independently → Deploy/Demo (Users can add items)
4. Add User Story 3 → Test independently → Deploy/Demo (Full quantity tracking)
5. Add User Story 4 → Test independently → Deploy/Demo (Enhanced audit trail)
6. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (View Inventory)
   - Developer B: User Story 2 (Add Items) - can work in parallel
   - Developer C: User Story 3 (Quantity Changes) - can work in parallel
3. Developer D starts User Story 4 after US1 is complete (extends transaction history)
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies - safe to run in parallel
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run migrations (T086) before deploying to any environment
- Ensure HTTPS is properly configured (nginx) for Telegram WebApp requirements (FR-008)
- Photo compression must meet targets: <500KB full-size, <50KB thumbnails (FR-003)
- All endpoints require Telegram WebApp authentication (FR-008)
- Optimistic locking prevents concurrent update issues (FR-007)
- Soft deletes preserve audit trail (FR-013)
- Automated backups required per Constitution Principle III (T076)

---

## Task Count Summary

- **Phase 1 (Setup)**: 8 tasks
- **Phase 2 (Foundational)**: 17 tasks
- **Phase 3 (User Story 1 - View Inventory)**: 14 tasks
- **Phase 4 (User Story 2 - Add Items)**: 14 tasks
- **Phase 5 (User Story 3 - Quantity Changes)**: 13 tasks
- **Phase 6 (User Story 4 - Transaction History)**: 6 tasks
- **Phase 7 (Polish & Cross-Cutting)**: 18 tasks

**Total Tasks**: 90 tasks

**MVP Scope** (Setup + Foundational + US1): 39 tasks
**Parallel Opportunities**: 45 tasks marked [P] can run in parallel within their phase
**Independent Test Criteria**: Each of 4 user stories has clear independent validation criteria
