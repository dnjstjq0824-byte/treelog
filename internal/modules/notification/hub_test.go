package notification

import (
	"testing"
	"time"
)

func TestHubPublishSubscribe(t *testing.T) {
	hub := NewHub()
	alice := make(chan Notification, 4)
	bob := make(chan Notification, 4)
	hub.Subscribe("alice", alice)
	hub.Subscribe("bob", bob)
	defer hub.Unsubscribe("alice", alice)
	defer hub.Unsubscribe("bob", bob)

	if got := hub.SubscriberCount("alice"); got != 1 {
		t.Fatalf("alice subscribers = %d", got)
	}
	item := Notification{ID: "n1", Category: "COMMENT", Title: "hi"}
	hub.Publish("alice", item)

	select {
	case got := <-alice:
		if got.ID != "n1" {
			t.Fatalf("alice got %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("alice received nothing")
	}
	select {
	case got := <-bob:
		t.Fatalf("bob must not receive alice's notification, got %+v", got)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHubUnsubscribeReleases(t *testing.T) {
	hub := NewHub()
	ch := make(chan Notification, 1)
	hub.Subscribe("u", ch)
	hub.Unsubscribe("u", ch)
	if got := hub.SubscriberCount("u"); got != 0 {
		t.Fatalf("subscribers after unsubscribe = %d", got)
	}
	// Publishing to nobody must not block or panic.
	hub.Publish("u", Notification{ID: "x"})
}

func TestHubSlowReaderDrops(t *testing.T) {
	hub := NewHub()
	ch := make(chan Notification) // unbuffered, nobody reading
	hub.Subscribe("u", ch)
	defer hub.Unsubscribe("u", ch)
	done := make(chan struct{})
	go func() {
		hub.Publish("u", Notification{ID: "x"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("publish blocked on a slow reader")
	}
}
