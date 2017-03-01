package events

import (
	"context"
	"time"
)

type EventTimerConfig struct {
	ctx  context.Context
	rx   chan Event
	tick time.Duration
	name string
}

func NewEventTimeout(cfg *EventTimerConfig) {
	go func() {
		timeout := time.After(cfg.tick)
		select {
		case <-cfg.ctx.Done():
			return
		case <-timeout:
			cfg.rx <- Event{Code: TimerExpired, Source: cfg.name}
		}
	}()
}

func NewEventTimer(cfg *EventTimerConfig) {
	go func() {
		ticker := time.NewTicker(cfg.tick)
		select {
		case <-cfg.ctx.Done():
			return
		case <-ticker.C:
			cfg.rx <- Event{Code: TimerExpired, Source: cfg.name}
		}
	}()
}
