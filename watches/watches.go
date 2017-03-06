package watches

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/joyent/containerpilot/commands"
	"github.com/joyent/containerpilot/discovery"
	"github.com/joyent/containerpilot/events"
)

// Watch represents a task to execute when something changes
type Watch struct {
	ID               string
	Tag              string
	exec             *commands.Command
	discoveryService discovery.ServiceBackend

	// Event handling
	events.EventHandler
	startupEvent   events.Event
	startupTimeout int
	poll           int
}

func NewWatch(cfg *WatchConfig) (*Watch, error) {
	watch := &Watch{}
	watch.ID = cfg.Name
	watch.poll = cfg.Poll
	watch.Tag = cfg.Tag

	watch.Rx = make(chan events.Event, 1000)
	watch.Flush = make(chan bool)
	watch.startupEvent = events.Event{Code: events.StatusChanged, Source: watch.ID}
	watch.startupTimeout = -1

	cmd, err := commands.NewCommand(cfg.OnChangeExec, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("could not parse `onChange` in watch %s: %s",
			cfg.Name, err)
	}
	watch.exec = cmd
	return watch, nil
}

// CheckForUpstreamChanges checks the service discovery endpoint for any changes
// in a dependent backend. Returns true when there has been a change.
func (watch *Watch) CheckForUpstreamChanges() bool {
	return watch.discoveryService.CheckForUpstreamChanges(watch.ID, watch.Tag)
}

// OnChange runs the watch's executable, returning an error on failure.
func (watch *Watch) OnChange(ctx context.Context) error {
	// TODO: we want to update Run... functions to accept
	// a parent context so we can cancel them from this main loop
	return commands.RunWithTimeout(watch.exec, log.Fields{
		"process": watch.startupEvent.Code, "watch": watch.ID})
}

func (watch *Watch) Run(bus *events.EventBus) {
	watch.Bus = bus
	ctx, cancel := context.WithCancel(context.Background())

	timerSource := fmt.Sprintf("%s-watch-timer", watch.ID)
	events.NewEventTimer(ctx, watch.Rx,
		time.Duration(watch.poll)*time.Second, timerSource)

	go func() {
		select {
		case event := <-watch.Rx:
			switch event.Code {
			case events.TimerExpired:
				if event.Source == timerSource {
					changed := watch.CheckForUpstreamChanges()
					if changed {
						watch.Bus.Publish(
							events.Event{Code: events.StatusChanged, Source: watch.ID})
					}
				}
			case events.Quit:
				if event.Source != watch.ID {
					break
				}
				fallthrough
			case events.Shutdown:
				watch.Unsubscribe(watch.Bus)
				close(watch.Rx)
				cancel()
				watch.Flush <- true
				return
			case watch.startupEvent.Code:
				watch.Bus.Publish(
					events.Event{Code: events.Started, Source: watch.ID})
				err := watch.OnChange(ctx)
				if err != nil {
					watch.Bus.Publish(
						events.Event{Code: events.ExitSuccess, Source: watch.ID})
				} else {
					watch.Bus.Publish(
						events.Event{Code: events.ExitSuccess, Source: watch.ID})
				}
			}
		}
	}()
}
