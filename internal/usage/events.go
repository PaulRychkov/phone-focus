package usage

import (
	"strconv"
	"strings"
)

const (
	codeActivityResumed      = 1
	codeActivityPaused       = 2
	codeScreenInteractive    = 15
	codeScreenNonInteractive = 16
	codeKeyguardShown        = 17
	codeKeyguardHidden       = 18
	codeActivityStopped      = 23
	codeDeviceShutdown       = 26
)

func eventName(code int) string {
	switch code {
	case codeActivityResumed:
		return eventActivityResumed
	case codeActivityPaused:
		return eventActivityPaused
	case codeActivityStopped:
		return eventActivityStopped
	case codeScreenInteractive:
		return eventScreenOn
	case codeScreenNonInteractive:
		return eventScreenOff
	case codeKeyguardShown:
		return eventKeyguardShown
	case codeKeyguardHidden:
		return eventKeyguardHidden
	case codeDeviceShutdown:
		return eventDeviceShutdown
	}
	return ""
}

func ParseEvents(raw string) []Event {
	lines := strings.Split(raw, "\n")
	events := make([]Event, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		timestamp, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			continue
		}
		code, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}
		name := eventName(code)
		if name == "" {
			continue
		}
		pkg := ""
		if len(parts) > 2 {
			pkg = strings.TrimSpace(parts[2])
		}
		events = append(events, Event{Time: timestamp, Type: name, Package: pkg})
	}
	return events
}
