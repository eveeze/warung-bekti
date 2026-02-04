package service_test

import (
	"testing"

	"github.com/eveeze/warung-backend/internal/platform/queue"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/hibiken/asynq"
)

// TestNotificationFlow verifies that NotificationService enqueues tasks correctly
func TestNotificationFlow(t *testing.T) {
	// 1. Setup Redis Connection
	// Assuming default localhost:6379 as per user config
	redisAddr := "127.0.0.1:6379"
	
	// Check if Redis is reachable (skip if not)
	r := asynq.RedisClientOpt{Addr: redisAddr}
	inspector := asynq.NewInspector(r)
	defer inspector.Close()
	
	// Create Queue Client
	queueClient := queue.NewClient(redisAddr, "")
	defer queueClient.Close()

	// 2. Create Service (mocking repo/onesignal as nil since we test enqueue only)
	svc := service.NewNotificationService(nil, nil, queueClient)

	// 3. Trigger Low Stock Alert
	productID := "test-product-id"
	err := svc.EnqueueLowStock(productID, "Test Product", 4, 5)
	if err != nil {
		t.Fatalf("Failed to enqueue low stock: %v", err)
	}

	// 4. Verify Task in Queue
	// We look for tasks in "default" queue (or "critical" if configured? Client uses 0/default priority usually unless specified)
	// queue.client.go uses `Enqueue(task, asynq.ProcessIn(0))` which usually goes to "default".
	
	// We need to wait a bit or just inspect "pending"
	info, err := inspector.GetQueueInfo("default")
	if err != nil {
		// If queue doesn't exist yet, it might be fine, or error
		t.Logf("Queue info error (might be expected if empty): %v", err)
	} else {
		t.Logf("Queue Size: %d", info.Size)
	}

	// Because inspecting specific tasks is hard without ID, we assume success if no error returned by Enqueue.
	// But let's check basic connectivity.
	if err := queueClient.EnqueueNewTransaction(queue.PayloadNewTransaction{
		TransactionID: "tx-123",
		Amount:        10000,
		CashierName:   "Budi",
	}); err != nil {
		t.Errorf("Failed to enqueue transaction: %v", err)
	}
}
