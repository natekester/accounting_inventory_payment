# CLAUDE.md - Development & TDD Guidelines

This repository follows strict **Test-Driven Development (TDD)** practices and adheres to the architecture outlined in [`design-doc.md`](./design-doc.md).

---

## 1. Test-Driven Development (TDD) Requirements

All features, bug fixes, and refactors MUST follow the **Red-Green-Refactor** workflow:

1. **Red (Write Failing Tests First)**:
   - Before writing any domain logic, API handler, repository method, or strategy implementation, write a unit or integration test that asserts the expected behavior.
   - Run the test suite to confirm that the test fails for the expected reason.
2. **Green (Write Minimal Implementation Code)**:
   - Implement only enough code to make the failing test pass cleanly.
   - Do NOT add extraneous features or unrequested abstractions during this step.
3. **Refactor**:
   - Clean up code, enforce domain boundaries, optimize performance, and ensure clean adherence to pattern guidelines without breaking passing tests.

### Verification Rules
* Never mark a task or feature complete without running the relevant test suites and verifying green test execution.
* Do NOT comment out failing assertions or swallow errors to pass tests.

---

## 2. Architecture & Design Patterns Reference

Always refer to [`design-doc.md`](./design-doc.md) for full architectural specifications.

### Key Architectural Rules
* **Modular Monolith & Schema Isolation**:
  * All modules (`inventory`, `payment`, `accounting`) share a single database instance (SQLite / Postgres), but each module owns a **dedicated database schema** (e.g. `inventory`, `payment`, `accounting`).
  * Direct cross-schema queries, joins, or DB mutations are strictly prohibited.
  * Modules communicate **exclusively via public Go APIs or Domain Events**.
* **Strategy Pattern**:
  * Vendor integrations (`StripeStrategy`, `RilletStrategy`) implement domain strategy interfaces (Ports).
  * Domain services (`PaymentService`, `AccountingService`) act as **Strategy Contexts**, selecting and executing strategies at runtime based on context parameters.
* **Database Abstraction**:
  * All database operations use **GORM**, allowing seamless transition from SQLite to PostgreSQL by altering configuration environment variables.

---

## 3. Testing Patterns & Frameworks

Refer to Section 4 of [`design-doc.md`](./design-doc.md#4-testing-patterns) for full details:

* **Unit Tests**: Domain services in Go using mocked repository and strategy interfaces (`mockery`).
* **Integration Tests**:
  * **Repository**: Tested using GORM with an in-memory SQLite connection (`file::memory:?cache=shared`).
  * **Gateways & API Controllers**: Gin router endpoints tested with Go's `net/http/httptest`.
* **True End-to-End (E2E) Tests**: **Playwright** driving the Next.js UI in headless browsers to validate full stack flows (UI → Gin Backend → Database).

---

## 4. Common Commands

### Backend (Go)
```bash
# Run all Go tests
cd backend && go test ./...

# Run unit tests only
cd backend && go test -v ./internal/modules/...

# Run with race detector
cd backend && go test -race ./...
```

### Frontend (Next.js & Playwright)
```bash
# Run Playwright E2E tests
cd frontend && npx playwright test

# Run Playwright UI mode
cd frontend && npx playwright test --ui
```

### Docker
```bash
# Build and spin up all containers
docker-compose up --build

# Run in background
docker-compose up -d
```
