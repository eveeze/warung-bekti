package handler

import (
	"fmt"
	"net/http"

	"github.com/eveeze/warung-backend/internal/service"
)

type EventHandler struct {
	eventSvc *service.EventService
}

func NewEventHandler(eventSvc *service.EventService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc}
}

// Events handles SSE connections
func (h *EventHandler) Events(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Flush headers immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Subscribe to events
	ch := h.eventSvc.Subscribe()
	defer h.eventSvc.Unsubscribe(ch)

	// Listen for connection close
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			// Send event
			// Format:
			// event: type
			// data: ...
			// \n\n
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", event.Data)
			flusher.Flush()
		}
	}
}
