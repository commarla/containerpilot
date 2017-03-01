package events

import (
	"context"
	"fmt"
	"time"

	"github.com/joyent/containerpilot/commands"
)

type Watch struct {
	Args    []string
	Command *commands.Command
	ID      string
	Poll    int

	// Event handling
	EventHandler
	startupEvent   Event
	startupTimeout int
	heartbeat      int
}

func (watch *Watch) Run() {
	// TODO: this will probably be a background context b/c we've got
	// message-passing to the main loop for cancellation
	ctx, cancel := context.WithCancel(context.TODO())

	timerSource := fmt.Sprintf("%s-watch-timer", watch.ID)
	timercfg := &EventTimerConfig{
		ctx:  ctx,
		rx:   watch.rx,
		tick: time.Duration(watch.heartbeat) * time.Second,
		name: timerSource,
	}
	NewEventTimer(timercfg)

	go func() {
		select {
		case event := <-watch.rx:
			switch event.Code {
			case TimerExpired:
				if event.Source == timerSource {
					fmt.Printf("checking: %s\n", watch.ID)
					watch.bus.Publish(Event{Code: StatusChanged, Source: watch.ID})
				}
			case Quit:
				if event.Source != watch.ID {
					break
				}
				fallthrough
			case Shutdown:
				watch.Unsubscribe(watch.bus)
				close(watch.rx)
				cancel()
				watch.flush <- true
				return
			case watch.startupEvent.Code:
				// run this in a goroutine and pass it our context
				watch.bus.Publish(Event{Code: Started, Source: watch.ID})
				fmt.Println("watch exec running!")
				watch.bus.Publish(Event{Code: ExitSuccess, Source: watch.ID})
			default:
				fmt.Println("don't care about this message")
			}
		}
	}()
}
