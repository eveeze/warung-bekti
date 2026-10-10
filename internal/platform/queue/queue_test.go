package queue

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestQueueClientAndServer(t *testing.T) {
	// Connect to local dev Redis on 127.0.0.1:6381
	redisAddr := "127.0.0.1:6381"
	client := NewClient(redisAddr, "")
	defer client.Close()

	// Enqueue tasks
	err := client.EnqueueLowStockAlert(PayloadLowStock{
		ProductID:    "p1",
		ProductName:  "Test Item",
		CurrentStock: 2,
		MinStock:     5,
	})
	if err != nil {
		t.Logf("enqueue low stock alert: %v", err)
	}

	err = client.EnqueueNewTransaction(PayloadNewTransaction{
		TransactionID: "tx1",
		Amount:        50000,
		CashierName:   "Budi",
	})
	if err != nil {
		t.Logf("enqueue new transaction: %v", err)
	}

	err = client.EnqueueSendNotification(PayloadNotificationSend{
		Title:   "Promo",
		Message: "Diskon 10%",
	})
	if err != nil {
		t.Logf("enqueue send notification: %v", err)
	}

	// Server test
	srv := NewServer(redisAddr, "", 2)
	srv.Handle(TypeLowStockAlert, func(ctx context.Context, task *asynq.Task) error {
		return nil
	})
	srv.Handle(TypeNewTransaction, func(ctx context.Context, task *asynq.Task) error {
		return nil
	})

	if err := srv.Start(); err != nil {
		t.Logf("server start: %v", err)
	} else {
		srv.Stop()
	}

	// Test Run
	srv2 := NewServer(redisAddr, "", 2)
	go func() {
		_ = srv2.Run()
	}()
	time.Sleep(10 * time.Millisecond)
	srv2.Stop()
}
