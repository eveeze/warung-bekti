package handler

import (
	"net/http"
	"strconv"

	"github.com/eveeze/warung-backend/internal/middleware"
	"github.com/eveeze/warung-backend/internal/pkg/response"
	"github.com/eveeze/warung-backend/internal/service"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationSvc *service.NotificationService
}

func NewNotificationHandler(notificationSvc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationSvc: notificationSvc,
	}
}

// GetNotifications retrieves notification history for the authenticated user
// GET /api/v1/notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Parse pagination params
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notificationSvc.GetUserNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch notifications")
		return
	}

	response.OK(w, "Notifications retrieved", map[string]interface{}{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
	})
}

// MarkAsRead marks a single notification as read
// PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Get notification ID from URL
	notifIDStr := r.PathValue("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid notification ID")
		return
	}

	// Mark as read
	if err := h.notificationSvc.MarkAsRead(r.Context(), notifID); err != nil {
		response.InternalServerError(w, "Failed to mark notification as read")
		return
	}

	response.OK(w, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications for the user as read
// PATCH /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	if err := h.notificationSvc.MarkAllAsRead(r.Context(), userID); err != nil {
		response.InternalServerError(w, "Failed to mark all notifications as read")
		return
	}

	response.OK(w, "All notifications marked as read", nil)
}
