package services

import (
	"context"
	"fmt"
	"time"

	"github.com/joyent/containerpilot/commands"
	"github.com/joyent/containerpilot/discovery"
	"github.com/joyent/containerpilot/events"
)

// Some magic numbers used internally by restart limits
const (
	haltRestarts      = -1
	unlimitedRestarts = -2
)

// Service configures the service discovery data
type Service struct {
	Name             string
	ID               string
	exec             *commands.Command
	heartbeat        int
	Status           bool // TODO: we'll need this to carry more info than bool
	discoveryService discovery.ServiceBackend
	Definition       *discovery.ServiceDefinition

	// Event handling
	events.EventHandler
	startupEvent   events.Event
	startupTimeout int
	restart        bool
	restartLimit   int
	restartsRemain int
	runEvery       int
}

// NewService creates a new service
func NewService(cfg *ServiceConfig) (*Service, error) {
	service := &Service{
		Name:             cfg.Name,
		ID:               cfg.ID,
		heartbeat:        cfg.Heartbeat,
		discoveryService: cfg.discoveryService,
		Definition:       cfg.definition,
	}
	if cfg.Exec != nil {
		cmd, err := commands.NewCommand(cfg.Exec, cfg.execTimeout)
		if err != nil {
			return nil, fmt.Errorf("could not parse `health` in service %s: %s", cfg.Name, err)
		}
		cmd.Name = fmt.Sprintf("%s.health", cfg.Name)
		service.exec = cmd
	}
	return service, nil
}

// SendHeartbeat sends a heartbeat for this service
func (svc *Service) SendHeartbeat() {
	svc.discoveryService.SendHeartbeat(svc.Definition)
}

// MarkForMaintenance marks this service for maintenance
func (svc *Service) MarkForMaintenance() {
	svc.discoveryService.MarkForMaintenance(svc.Definition)
}

// Deregister will deregister this instance of the service
func (svc *Service) Deregister() {
	svc.discoveryService.Deregister(svc.Definition)
}

func (svc *Service) Run(bus *events.EventBus) {
	svc.Bus = bus
	ctx, cancel := context.WithCancel(context.Background())

	runEverySource := fmt.Sprintf("%s-run-every", svc.ID)
	if svc.runEvery > 0 {
		events.NewEventTimeout(ctx, svc.Rx,
			time.Duration(svc.runEvery)*time.Second, runEverySource)
	}

	heartbeatSource := fmt.Sprintf("%s-heartbeat", svc.ID)
	if svc.heartbeat > 0 {
		events.NewEventTimeout(ctx, svc.Rx,
			time.Duration(svc.heartbeat)*time.Second, heartbeatSource)
	}

	timeoutSource := fmt.Sprintf("%s-wait-timeout", svc.ID)
	if svc.startupTimeout > 0 {
		events.NewEventTimeout(ctx, svc.Rx,
			time.Duration(svc.startupTimeout)*time.Second, timeoutSource)
	}

	go func() {
		select {
		case event := <-svc.Rx:
			switch event.Code {
			case events.TimerExpired:
				switch event.Source {
				case heartbeatSource:
					// non-advertised services shouldn't receive this event
					// but we'll hit a null-pointer if we screw it up
					if svc.Status == true && svc.Definition != nil {
						svc.SendHeartbeat()
					}
				case timeoutSource:
					svc.Bus.Publish(events.Event{Code: events.TimerExpired, Source: svc.ID})
					svc.Rx <- events.Event{Code: events.Quit, Source: svc.ID}
				case runEverySource:
					if !svc.restart || (svc.restartLimit != unlimitedRestarts &&
						svc.restartsRemain <= haltRestarts) {
						break
					}
					svc.restartsRemain--
					svc.Rx <- events.Event{Code: svc.startupEvent.Code, Source: svc.ID}
				}
			case events.Quit:
				if event.Source == svc.ID {
					break
				}
				fallthrough
			case events.Shutdown:
				svc.Unsubscribe(svc.Bus)
				close(svc.Rx)
				cancel()
				svc.Flush <- true
				return
			case events.ExitSuccess:
			case events.ExitFailed:
				if event.Source != svc.ID {
					break
				}
				if !svc.restart || (svc.restartLimit != unlimitedRestarts &&
					svc.restartsRemain <= haltRestarts) {
					break
				}
				svc.restartsRemain--
				svc.Rx <- events.Event{Code: svc.startupEvent.Code, Source: svc.ID}
			case svc.startupEvent.Code:
				// run this in a goroutine and pass it our context
				svc.Bus.Publish(events.Event{Code: events.Started, Source: svc.ID})
				fmt.Println("running!")
				svc.Bus.Publish(events.Event{Code: events.ExitSuccess, Source: svc.ID})
			}
		}
	}()
}
