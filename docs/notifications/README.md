# Notification System

The Notification System handles asynchronous background tasks and user alerts (push notifications, email, etc.) using a reliable queue-based architecture.

## Architecture

We use **[hibiken/asynq](https://github.com/hibiken/asynq)** for the job queue system, backed by **Redis**. This ensures task persistence and reliability across server restarts.

### Components

1.  **Queue Client (`internal/platform/queue/client.go`)**:
    - Used by services (e.g., `TransactionService`) to enqueue tasks.
    - Responsbile for serializing payloads.

2.  **Queue Server (`internal/platform/queue/server.go`)**:
    - Runs in the background (started in `main.go`).
    - Consumes tasks from Redis and executes registered handlers.
    - Handles retries and concurrency.

3.  **Notification Service (`internal/service/notification_svc.go`)**:
    - Core business logic for handling notification tasks.
    - Integrates with `OneSignal` for external push notifications.
    - Saves notification history to the PostgreSQL database (`notifications` table).

4.  **OneSignal Client (`internal/integration/onesignal/client.go`)**:
    - Wrapper for the OneSignal REST API.

## Configuration

The system requires the following environment variables in `.env`:

```env
# Redis (Required for Queue)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# OneSignal (Optional - for Push Notifications)
ONESIGNAL_APP_ID=your-app-id
ONESIGNAL_API_KEY=your-api-key
```

## Supported Tasks

### 1. Low Stock Alert (`notification:low_stock`)

- **Trigger**: When a transaction reduces a product's stock below its `min_stock_alert` threshold.
- **Action**:
  - Saves a "system" notification to the DB.
  - Sends a push notification via OneSignal (if configured).

### 2. New Transaction (`notification:new_transaction`)

- **Trigger**: When a transaction is successfully completed.
- **Action**: (Currently logs the transaction, extendable to admin alerts).

## Usage

### Enqueuing a Task

In your service, inject `*queue.Client` or `*service.NotificationService`.

```go
// Direct Queue Enqueue
client.EnqueueLowStockAlert(queue.PayloadLowStock{...})

// OR via Service (Preferred wrapper)
notificationSvc.EnqueueLowStock(productID, name, currentStock, minStock)
```

### Retrieving Notifications

Use the HTTP API to fetch notification history for users.

**Endpoint**: `GET /api/v1/notifications` (Not yet exposed publicly, need to add handler if frontend needs it)

## Database Schema

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID, -- NULL for system/global
    title VARCHAR(255),
    message TEXT,
    type VARCHAR(50),
    data JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP
);
```
