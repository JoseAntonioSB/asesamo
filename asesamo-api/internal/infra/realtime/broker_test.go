package realtime_test

import (
	"sync"
	"testing"
	"time"

	"asesamo-api/internal/infra/realtime"
)

func TestEventBroker_SingleTenantBroadcast(t *testing.T) {
	broker := realtime.NewInMemoryEventBroker(10)
	tenantID := "11111111-1111-1111-1111-111111111111"

	ch, unsubscribe := broker.Subscribe(tenantID)
	defer unsubscribe()

	event := realtime.Event{
		Type:      "evacuation.started",
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"evacuation_number": "EVAC-2026-0001",
			"evacuation_type":   "fire",
		},
	}

	broker.Publish(event)

	select {
	case received := <-ch:
		if received.Type != "evacuation.started" {
			t.Fatalf("expected evacuation.started, got %s", received.Type)
		}
		if received.Payload["evacuation_number"] != "EVAC-2026-0001" {
			t.Fatalf("unexpected payload: %v", received.Payload)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestEventBroker_TenantIsolation(t *testing.T) {
	broker := realtime.NewInMemoryEventBroker(10)
	tenantA := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantB := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	chA, unsubA := broker.Subscribe(tenantA)
	defer unsubA()

	chB, unsubB := broker.Subscribe(tenantB)
	defer unsubB()

	broker.Publish(realtime.Event{
		Type:      "alarm.triggered",
		TenantID:  tenantA,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"level": "critical"},
	})

	// chA should receive it
	select {
	case received := <-chA:
		if received.Type != "alarm.triggered" {
			t.Fatalf("expected alarm.triggered, got %s", received.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("chA should have received the event")
	}

	// chB should NOT receive it
	select {
	case unexpected := <-chB:
		t.Fatalf("chB should not receive event from tenantA: %v", unexpected)
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	}
}

func TestEventBroker_Unsubscribe(t *testing.T) {
	broker := realtime.NewInMemoryEventBroker(10)
	tenantID := "11111111-1111-1111-1111-111111111111"

	ch, unsubscribe := broker.Subscribe(tenantID)
	unsubscribe()

	broker.Publish(realtime.Event{
		Type:      "test.event",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})

	select {
	case ev, ok := <-ch:
		if ok {
			t.Fatalf("unexpected event after unsubscribe: %v", ev)
		}
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestEventBroker_ConcurrentSafety(t *testing.T) {
	broker := realtime.NewInMemoryEventBroker(50)
	tenantID := "tenant-concurrent-test"

	var wg sync.WaitGroup
	subscriberCount := 10
	publishCount := 50

	var subs []func()
	var channels []<-chan realtime.Event

	for i := 0; i < subscriberCount; i++ {
		ch, unsub := broker.Subscribe(tenantID)
		channels = append(channels, ch)
		subs = append(subs, unsub)
	}

	// Start reading goroutines
	var receivedCounts sync.Map
	for idx, ch := range channels {
		wg.Add(1)
		go func(subIdx int, c <-chan realtime.Event) {
			defer wg.Done()
			count := 0
			for range c {
				count++
				if count == publishCount {
					receivedCounts.Store(subIdx, count)
					return
				}
			}
		}(idx, ch)
	}

	// Publish concurrently
	for i := 0; i < publishCount; i++ {
		go func(n int) {
			broker.Publish(realtime.Event{
				Type:      "concurrent.event",
				TenantID:  tenantID,
				Timestamp: time.Now(),
				Payload:   map[string]interface{}{"index": n},
			})
		}(i)
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for all concurrent events to be received")
	}

	for _, unsub := range subs {
		unsub()
	}
}
