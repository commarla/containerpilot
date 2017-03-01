package events

import (
	"testing"
	"time"
)

func TestServiceEvents(t *testing.T) {

	svc := Service{ID: "myservice"}
	svc.bus = NewEventBus()
	svc.rx = make(chan Event, 1000)
	svc.flush = make(chan bool)
	svc.startupEvent = Event{Code: ExitSuccess, Source: "upstream"}
	svc.startupTimeout = 60
	svc.restarts = 0
	svc.heartbeat = 3

	svc.Run()
	svc.bus.Publish(Event{Code: Started, Name: "serviceA"})

	svc.Close()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked but should not: sent to closed Subscriber")
		}
	}()
	svc.bus.Publish(Event{Code: Started, Name: "serviceA"}) // should not panic
}

func TestServiceTimeout(t *testing.T) {

	svc := Service{ID: "myservice"}
	svc.bus = NewEventBus()
	svc.rx = make(chan Event, 1000)
	svc.flush = make(chan bool)
	svc.startupEvent = Event{Code: Startup}
	svc.startupTimeout = 1
	svc.restarts = 0
	svc.heartbeat = 3

	svc.Run()
	svc.bus.Publish(Event{Code: Started, Name: "serviceA"})

	// note that we can't send a .Close() after this b/c we've timed out
	// and we'll end up blocking forever
	time.Sleep(1 * time.Second)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panicked but should not: sent to closed Subscriber")
		}
	}()
	svc.bus.Publish(Event{Code: Started, Name: "serviceA"}) // should not panic
}
