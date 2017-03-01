package events

import (
	"context"
	"fmt"
	"time"

	"github.com/joyent/containerpilot/commands"
)

type HealthCheck struct {
	Args    []string
	Command *commands.Command
	ID      string
	Poll    int

	// Event handling
	EventHandler
	startupEvent   Event
	startupTimeout int
	restarts       int
	heartbeat      int
}

func (check *HealthCheck) Run() {
	// TODO: this will probably be a background context b/c we've got
	// message-passing to the main loop for cancellation
	ctx, cancel := context.WithCancel(context.TODO())

	timerSource := fmt.Sprintf("%s-check-timer", check.ID)
	timercfg := &EventTimerConfig{
		ctx:  ctx,
		rx:   check.rx,
		tick: time.Duration(check.heartbeat) * time.Second,
		name: timerSource,
	}
	NewEventTimer(timercfg)

	go func() {
		select {
		case event := <-check.rx:
			switch event.Code {
			case TimerExpired:
				if event.Source == timerSource {
					fmt.Printf("checking: %s\n", check.ID)
					check.bus.Publish(Event{Code: StatusChanged, Source: check.ID})
				}
			case Quit:
				if event.Source == check.ID {
					break
				}
				fallthrough
			case Shutdown:
				check.Unsubscribe(check.bus)
				close(check.rx)
				cancel()
				check.flush <- true
				return
			case check.startupEvent.Code:
				// run this in a goroutine and pass it our context
				check.bus.Publish(Event{Code: Started, Source: check.ID})
				fmt.Println("check exec running!")
				check.bus.Publish(Event{Code: ExitSuccess, Source: check.ID})
			default:
				fmt.Println("don't care about this message")
			}
		}
	}()
}
