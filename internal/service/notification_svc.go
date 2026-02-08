package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/eveeze/warung-backend/internal/integration/onesignal"
	"github.com/eveeze/warung-backend/internal/platform/queue"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type NotificationService struct {
	repo         *repository.NotificationRepository
	oneSignal    *onesignal.Client
	queueClient  *queue.Client // For re-enqueueing if needed, or initiating
}

func NewNotificationService(repo *repository.NotificationRepository, oneSignal *onesignal.Client, queueClient *queue.Client) *NotificationService {
	return &NotificationService{
		repo:        repo,
		oneSignal:   oneSignal,
		queueClient: queueClient,
	}
}

// EnqueueLowStock enqueues a low stock alert
func (s *NotificationService) EnqueueLowStock(productID string, productName string, currentStock, minStock int) error {
	return s.queueClient.EnqueueLowStockAlert(queue.PayloadLowStock{
		ProductID:    productID,
		ProductName:  productName,
		CurrentStock: currentStock,
		MinStock:     minStock,
	})
}

// Handlers for Asynq

func (s *NotificationService) HandleLowStockTask(ctx context.Context, t *asynq.Task) error {
	var p queue.PayloadLowStock
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	title := "Low Stock Alert"
	message := fmt.Sprintf("Product %s is running low (%d left). Reorder soon!", p.ProductName, p.CurrentStock)
	
	// 1. Save to DB (System notification)
	data, _ := json.Marshal(map[string]interface{}{
		"product_id": p.ProductID,
		"current_stock": p.CurrentStock,
	})
	
	notification := &repository.Notification{
		Title:   title,
		Message: message,
		Type:    "low_stock",
		Data:    data,
		IsRead:  false,
	}
	// Assuming system notifications have nil UserID
	if err := s.repo.Create(ctx, notification); err != nil {
		return fmt.Errorf("repo.Create failed: %w", err)
	}

	// 2. Send Push Notification (To all staff/admins)
	// In real app, you might fetch admin UserIDs here. keeping empty for All.
	if s.oneSignal != nil {
		// Only send if configured
		if err := s.oneSignal.SendNotification(title, message, nil, map[string]interface{}{
			"type": "low_stock",
			"product_id": p.ProductID,
		}); err != nil {
			log.Printf("Failed to send push: %v", err)
			// Don't fail the task just because push failed? 
			// Or maybe do fail so it retries? For now, log only.
		}
	}

	log.Printf("Processed Low Stock Alert for %s", p.ProductName)
	return nil
}

func (s *NotificationService) HandleNewTransactionTask(ctx context.Context, t *asynq.Task) error {
	var p queue.PayloadNewTransaction
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	
	// Logic for transaction notification...
	log.Printf("Processed Transaction %s from %s", p.TransactionID, p.CashierName)
	return nil
}

// GetUserNotifications returns notification history
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}
