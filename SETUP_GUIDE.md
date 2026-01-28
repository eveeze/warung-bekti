# WarungOS Backend - Setup & Running Guide

This guide covers how to start the backend services, the application itself, and how to verify everything using the provided Postman collection.

## 1. Prerequisites

Ensure you have the following installed:
- **Docker & Docker Compose** (for database, cache, storage)
- **Go 1.22+** (to run the application)
- **Postman** (for API testing)

## 2. Infrastructure Setup

We use Docker to run our dependencies (PostgreSQL, Redis, Minio).

### Option A: Clean Start (Recommended)
This command starts all infrastructure services in the background.

```bash
make infra-up
```

If you don't have `make` installed:
```bash
docker-compose up -d postgres redis minio
```

Wait about 10-15 seconds for the database to be fully ready.

## 3. Running the Application

You have two ways to run the backend API.

### Option A: Run Locally (Fastest)
This runs the Go app on your host machine, connecting to the Docker services.

1. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the Server:**
   ```bash
   go run ./cmd/api
   ```
   
   You should see logs indicating the server is running on port `:8080`.

### Option B: Run via Docker (Full Stack)
This runs the Go app inside a container as well.

```bash
make docker-up
```

## 4. API Testing with Postman

We have provided a comprehensive `warung-backend.postman_collection.json` file in the project root. This collection is "Enterprise Grade" with built-in tests and automated variable handling.

### How to Import
1. Open Postman.
2. Click **Import** -> **File** -> **Upload Files**.
3. Select `warung-backend.postman_collection.json` from this project folder.

### How to Run Automated Tests
This collection handles validation (Status 200/201), capturing IDs (Product ID, Customer ID) automatically, and reusing them in subsequent requests.

1. Click on the collection name **"WarungOS Backend API"** in the sidebar.
2. Click the **"Run"** button (Run Collection).
3. Ensure no environment is selected (the collection uses Collection Variables which are self-contained).
4. Click **"Run WarungOS Backend API"**.

### What Gets Tested?
- **Health Checks**: Verifies DB/Redis connections.
- **Products**: Creates products (One normal, one service-type), searches by barcode.
- **Inventory**: Restocks items, adjusts stock for damage.
- **Customers**: Creates normal customers and "bad credit" customers.
- **POS Transactions**:
  - Calculates bulk prices (Variable Pricing logic).
  - Processes Cash payments.
  - Processes Kasbon (Credit) payments.
  - **Edge Cases**: Attempts to buy more than available stock (expecting 400), attempts to exceed credit limit (expecting 400).
- **Kasbon**: Verifies debt history and records partial payments.
- **Reports**: Checks the daily dashboard and reports.

## 5. Deployment / Production

To build a production-ready Docker image:

```bash
# Build the minimal image (~15MB)
make docker-build
```

The resulting image uses a multi-stage build process for security and minimal size.

## Troubleshooting

- **Database Connection Error**: Ensure `make infra-up` was successful and wait a few seconds for Postgres to initialize.
- **Port 8080 Busy**: If another app is using port 8080, edit `.env` and `docker-compose.yml` to change `SERVER_PORT`.
