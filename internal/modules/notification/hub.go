package notification

import "sync"

// streamBuffer caps queued events per stream. Slow readers drop messages
// rather than blocking publishers or growing without bound.
const streamBuffer = 32

// Hub fans out newly created notifications to live SSE streams.
// It is process-local: deployments running multiple API replicas need a
// Redis pub/sub fanout to replace it.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan Notification]struct{}
}

// NewHub builds an empty Hub.
func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[chan Notification]struct{})}
}

// Subscribe registers ch for notifications addressed to userID. The caller
// must call Unsubscribe with the same channel to release it.
func (h *Hub) Subscribe(userID string, ch chan Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.subs[userID]
	if !ok {
		set = make(map[chan Notification]struct{})
		h.subs[userID] = set
	}
	set[ch] = struct{}{}
}

// Unsubscribe releases a channel previously registered for userID.
func (h *Hub) Unsubscribe(userID string, ch chan Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set, ok := h.subs[userID]
	if !ok {
		return
	}
	delete(set, ch)
	if len(set) == 0 {
		delete(h.subs, userID)
	}
}

// Publish delivers item to every stream subscribed for recipientID.
// Delivery never blocks: a full subscriber buffer drops the message.
func (h *Hub) Publish(recipientID string, item Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[recipientID] {
		select {
		case ch <- item:
		default:
		}
	}
}

// SubscriberCount reports live streams for userID (tests and operations).
func (h *Hub) SubscriberCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[userID])
}
