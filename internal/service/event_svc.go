package service

import (
	"encoding/json"
	"sync"
)

type EventType string

const (
	EventStockUpdate EventType = "stock_update"
)

type Event struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}

type EventService struct {
	clients map[chan Event]bool
	mu      sync.Mutex
}

func NewEventService() *EventService {
	return &EventService{
		clients: make(map[chan Event]bool),
	}
}

func (s *EventService) Subscribe() chan Event {
	ch := make(chan Event, 100) // Buffer to prevent blocking
	s.mu.Lock()
	s.clients[ch] = true
	s.mu.Unlock()
	return ch
}

func (s *EventService) Unsubscribe(ch chan Event) {
	s.mu.Lock()
	delete(s.clients, ch)
	s.mu.Unlock()
	close(ch)
}

func (s *EventService) Publish(eventType EventType, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	
	event := Event{
		Type: eventType,
		Data: bytes,
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- event:
		default:
			// Drop event if client is too slow (non-blocking)
		}
	}
}
