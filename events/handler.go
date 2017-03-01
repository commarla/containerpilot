package events

// EventHandler should be embedded in all Runners so that we can
// reuse the code for registering and unregistering handlers.
// This means we can't use struct literals for constructors and all
// NewRunner functions will need to set these fields explicitly:
//   runner.rx = make(chan Event, n)
//   runner.flush = make(chan bool)
//   runner.bus = &EventBus{}
type EventHandler struct {
	bus   *EventBus
	rx    chan Event // typically buffered
	flush chan bool  // must be unbuffered
}

// Subscribe adds the EventHandler to the list of handlers that
// receive all messages from the EventBus.
func (evh *EventHandler) Subscribe(bus *EventBus) {
	bus.Register(evh)
}

// Unsubscribe removes the EventHandler from the list of handlers
// that receive messages from the EventBus.
func (evh *EventHandler) Unsubscribe(bus *EventBus) {
	bus.Unregister(evh)
}

// Receive accepts an Event for the EventHandler's receive channel.
// Embedding struct should use a non-blocking buffered channel but
// this may be blocking in tests.
func (evh *EventHandler) Receive(e Event) {
	// TODO: instrument receives so we can report event throughput
	// statistics via Prometheus
	evh.rx <- e
}

// Close sends a Quit message to the EventHandler and then synchronously
// waits for the EventHandler to be unregistered from all events.
func (evh *EventHandler) Close() {
	evh.rx <- Event{Code: Quit}
	<-evh.flush
	close(evh.flush)
}
