# Modular Monolith: Inventory, Payments & Accounting

A modern, highly decoupled **Modular Monolith** application built with a **Go** backend and a **Next.js** frontend. It demonstrates clean architecture, hexagonal boundaries (Ports & Adapters), Strategy patterns, single-database multi-schema isolation, and event-driven integration.

---

## 🛠️ Tech Stack & Architecture

### Backend (Go)
* **Gin Web Framework:** REST API routing and HTTP controller request handling.
* **GORM ORM:** Database abstraction layer configured with a pure-Go SQLite driver (`glebarez/sqlite`) for rapid local development and PostgreSQL compatibility for production.
* **Hexagonal Architecture (Ports & Adapters):** Core domain logic is decoupled from external vendors using interface boundaries.
* **Strategy Pattern & Strategy Context:** Third-party integrations are treated as interchangeable strategies executed dynamically at runtime (e.g. `StripeStrategy` for payments, `RilletStrategy` for accounting).
* **Decoupled Module Schemas:** All modules (`inventory`, `payment`, `accounting`) reside in a single database connection but occupy **strictly isolated database schemas** (`inventory`, `payment`, `accounting` namespaces). Direct cross-schema queries or joins are prohibited.
* **Domain Events / EventBus:** In-memory pub-sub event bus used for side effects. For example, processing a successful payment publishes a `payment.completed` event, which the accounting module subscribes to and records in the journal ledger asynchronously.

### Frontend (Next.js)
* **Next.js App Router & TypeScript:** Sleek, modern responsive UI.
* **Vanilla CSS Glassmorphism Design:** Vibrant, premium look with zero external styling dependencies.
* **Live Features:**
  * 📦 **Inventory Management:** Create items, view stock counts, and adjust stock quantities.
  * 💳 **Payment Processing:** Trigger payments dynamically via `StripeStrategy`.
  * 📊 **Accounting Ledger:** Live bookkeeping logs synchronized automatically via backend domain events.

---

## 📂 Project Structure

```
├── backend/
│   ├── cmd/server/main.go            # Entry point for Gin application
│   ├── internal/
│   │   ├── shared/                   # Shared connection managers & EventBus
│   │   └── modules/                  # Isolated business domains
│   │       ├── inventory/            # Inventory Management module
│   │       ├── payment/              # Payments Module (StripeStrategy)
│   │       └── accounting/           # Accounting Module (RilletStrategy)
│   └── Dockerfile                    # Multi-stage dev/prod Docker configurations
├── frontend/
│   ├── src/app/                      # Next.js App components & styles
│   └── Dockerfile                    # Multi-stage dev/prod Docker configurations
├── docker-compose.yml                # Main multi-container dev/prod setup
├── design-doc.md                     # Comprehensive architecture design doc
└── CLAUDE.md                         # TDD requirements & developer guidelines
```

---

## 🚀 How to Run

### 1. Docker Compose (Recommended)
Build and spin up the entire frontend, backend (with Air live-reloading), and database stack:
```bash
docker-compose up --build
```
* Frontend Dashboard: [http://localhost:3000](http://localhost:3000)
* Backend Health Check: [http://localhost:8080/healthz](http://localhost:8080/healthz)

### 2. Local Setup
#### Backend:
```bash
cd backend
go run cmd/server/main.go
```
#### Frontend:
```bash
cd frontend
npm run dev
```

---

## 🧪 Testing Patterns

We follow strict **Test-Driven Development (TDD)** guidelines (documented in [`CLAUDE.md`](./CLAUDE.md)):

* **Unit Testing:** Write mock tests for services (`go test ./...` in `backend`).
* **Integration Testing:** Repositories tested using an in-memory SQLite database connection. Gin handlers tested via Go's `net/http/httptest`.
* **E2E Testing:** Full user workflow integration tests driven by **Playwright** against the browser UI.
