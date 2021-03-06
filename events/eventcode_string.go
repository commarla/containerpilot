// Code generated by "stringer -type EventCode"; DO NOT EDIT.

package events

import "fmt"

const eventCodename = "NoneExitSuccessExitFailedStoppingStoppedStatusHealthyStatusUnhealthyStatusChangedTimerExpiredEnterMaintenanceExitMaintenanceErrorQuitMetricStartupShutdownSignal"

var eventCodeindex = [...]uint8{0, 4, 15, 25, 33, 40, 53, 68, 81, 93, 109, 124, 129, 133, 139, 146, 154, 160}

func (i EventCode) String() string {
	if i < 0 || i >= EventCode(len(eventCodeindex)-1) {
		return fmt.Sprintf("EventCode(%d)", i)
	}
	return eventCodename[eventCodeindex[i]:eventCodeindex[i+1]]
}
