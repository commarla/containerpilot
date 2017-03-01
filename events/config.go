package events

/*
TODO: Service.Events will be the output of HealthChecks, OnChange handlers,
Command exits, Upstreams, and restarts. ex.

map[Event]action {
   Event{Code: Healthy, Name: "mybackend"}: func(s Service, e Event) error { return nil },
   Event{Code: Unhealthy, Name: "myotherbackend"}:  func(s Service, e Event) error { return nil },
   Event{Code: Success, Name: "myprestart"}:  func(s Service, e Event) error { return nil },
   Event{Code: Exited, Name: "myprestart"}:  func(s Service, e Event) error { return nil },
}

Example configuration
tasks:

  nginx:
    heartbeat: 3s
    ttl: 6s
    port: 80
    exec:
      start: nginx -g "daemon off;"

  app:
    heartbeat: 3s
    ttl: 6s
    port: 3000
    events:
      dbconnect:
        onSuccess:
          timeout: 60s
          exec: start
      db:
        onChange:
          exec: reload-db
    exec:
      start: node server.js
      reload-db: reload-db-connections.sh

  dbconnect:
    restart: never
    events:
      db:
        onHealthy:
          exec: configure-db
          timeout: 60s
    exec:
      configure-db: configure-db-connection.sh

func eventFromName(eventName, eventSource string) (Event, error) {
	var event Event
	switch eventName {
	case "onSuccess":
		event = Event{Code: ExitSuccess, Name: eventSource}
	case "onFail":
		event = Event{Code: ExitFailed, Name: eventSource}
	case "onHealthy":
		event = Event{Code: StatusHealthy, Name: eventSource}
	case "onUnhealthy":
		event = Event{Code: StatusUnhealthy, Name: eventSource}
	case "onChange":
		event = Event{Code: StatusChanged, Name: eventSource}
	default:
		return Event{}, fmt.Errorf("no event of type '%s'", eventName)
	}
	return event, nil
}

type configMap map[string]map[string]string

var eventsInput = map[string]interface{}{
	"events": map[string]interface{}{
		"eventSourceA": map[string]interface{}{
			"timeout": "60s",
			"exec":    "start",
		},
		"eventSourceB": map[string]interface{}{
			"exec": "reload-db",
		},
	},
}

var execInput = map[string]interface{}{
	"exec": map[string]interface{}{
		"start":     "run-my-command.sh",
		"reload-db": "reload-db-configuration.sh",
	},
}

eventSource:
  eventType:
	timeout: 60s
	exec: start
eventSource::
  eventType:
	exec: reload-db

func setupEvents(svc *Service, eventConfig configMap) (map[Event]action, error) {
	events := make(map[Event]action)
	for eventSource, typeActionPairs := range eventConfig {

		// we want these outside the loop b/c we might modify it
		var (
			event Event
			wait  bool
		)
		for eventType, eventAction := range typeActionPairs {

			// figure out what we're waiting for
			var (
				err error
			)
			switch eventType {
			case "wait":
				// 'wait' is not an event itself but a configuration option
				event, err = eventFromName(eventAction, eventSource)
				wait = true
			case "timeout":
				wait = true
				//				timeoutConfig = eventAction // the config overloads this value
			default:
				event, err = eventFromName(eventType, eventSource)

			}
			if err != nil {
				return nil, err
			}

			// pair the Event with the action
			switch eventAction {
			case "run":
			default:
				//				events[event] = runtask // TODO: this should return a new func where we create the task context
			}
		}
	}
	return events, nil
}

*/
