package events

import (
	"context"
	"fmt"
	"time"

	"github.com/joyent/containerpilot/commands"
)

/*
TODO: this is temporary while I hack out how this will interact
with everything else. It'll go in the `services` package when I'm done
*/

type Service struct {
	Args      []string
	Command   *commands.Command
	ID        string
	Heartbeat int
	Restart   string // TODO config
	Wait      int    // TODO config
	WaitOn    string // TODO config
	Status    bool   // TODO config

	// Event handling
	EventHandler
	startupEvent   Event
	startupTimeout int
	restarts       int // -1 for "always"
	runEvery       int
	heartbeat      int
}

func (svc *Service) Run() {
	// TODO: this will probably be a background context b/c we've got
	// message-passing to the main loop for cancellation
	ctx, cancel := context.WithCancel(context.TODO())

	runEverySource := fmt.Sprintf("%s-run-every", svc.ID)
	if svc.runEvery > 0 {
		timercfg := &EventTimerConfig{
			ctx:  ctx,
			rx:   svc.rx,
			tick: time.Duration(svc.runEvery) * time.Second,
			name: runEverySource,
		}
		NewEventTimeout(timercfg)
	}

	heartbeatSource := fmt.Sprintf("%s-heartbeat", svc.ID)
	if svc.heartbeat > 0 {
		timercfg := &EventTimerConfig{
			ctx:  ctx,
			rx:   svc.rx,
			tick: time.Duration(svc.heartbeat) * time.Second,
			name: heartbeatSource,
		}
		NewEventTimeout(timercfg)
	}
	timeoutSource := fmt.Sprintf("%s-wait-timeout", svc.ID)
	if svc.startupTimeout > 0 {
		timercfg := &EventTimerConfig{
			ctx:  ctx,
			rx:   svc.rx,
			tick: time.Duration(svc.startupTimeout) * time.Second,
			name: timeoutSource,
		}
		NewEventTimeout(timercfg)
	}

	go func() {
		select {
		case event := <-svc.rx:
			switch event.Code {
			case TimerExpired:
				switch event.Source {
				case heartbeatSource:
					if svc.Status == true {
						fmt.Printf("heartbeat: %s\n", svc.ID)
					}
				case timeoutSource:
					svc.bus.Publish(Event{Code: TimerExpired, Source: svc.ID})
					svc.rx <- Event{Code: Quit, Source: svc.ID}
				case runEverySource:
					if svc.restarts > 0 || svc.restarts < 0 {
						svc.rx <- Event{Code: svc.startupEvent.Code, Source: svc.ID}
						svc.restarts--
					}
				}
			case Quit:
				if event.Source == svc.ID {
					break
				}
				fallthrough
			case Shutdown:
				svc.Unsubscribe(svc.bus)
				close(svc.rx)
				cancel()
				svc.flush <- true
				return
			case ExitSuccess:
			case ExitFailed:
				if event.Source == svc.ID && svc.restarts > 0 || svc.restarts < 0 {
					svc.rx <- Event{Code: svc.startupEvent.Code, Source: svc.ID}
					svc.restarts--
				}
			case svc.startupEvent.Code:
				// run this in a goroutine and pass it our context
				svc.bus.Publish(Event{Code: Started, Source: svc.ID})
				fmt.Println("running!")
				svc.bus.Publish(Event{Code: ExitSuccess, Source: svc.ID})
			default:
				fmt.Println("don't care about this message")
			}
		}
	}()
}
