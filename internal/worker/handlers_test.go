package worker

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDefaultHandlersCoverEmittedEvents(t *testing.T) {
	handlers := DefaultHandlers(Dependencies{})
	for _, eventType := range []string{"system.ping", "post.published", "post.unpublished", "media.process"} {
		handler, ok := handlers[eventType]
		if !ok || handler == nil {
			t.Fatalf("missing handler for %s", eventType)
		}
	}
	if err := handlers["system.ping"](context.Background(), json.RawMessage(`{}`)); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestHandlersRejectInvalidPayloads(t *testing.T) {
	handlers := DefaultHandlers(Dependencies{})
	ctx := context.Background()
	cases := map[string]string{
		"post.published":   `{"post_id":""}`,
		"post.unpublished": `not-json`,
		"media.process":    `{}`,
	}
	for eventType, payload := range cases {
		if err := handlers[eventType](ctx, json.RawMessage(payload)); err == nil {
			t.Fatalf("%s accepted %q", eventType, payload)
		}
	}
}

func TestMediaProcessRequiresProcessor(t *testing.T) {
	handlers := DefaultHandlers(Dependencies{})
	err := handlers["media.process"](context.Background(), json.RawMessage(`{"asset_id":"a"}`))
	if err == nil {
		t.Fatalf("unconfigured media processor accepted work")
	}
}
