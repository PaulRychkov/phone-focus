package usage

import "sort"

const (
	eventActivityResumed = "ACTIVITY_RESUMED"
	eventActivityPaused  = "ACTIVITY_PAUSED"
	eventActivityStopped = "ACTIVITY_STOPPED"
	eventMoveForeground  = "MOVE_TO_FOREGROUND"
	eventMoveBackground  = "MOVE_TO_BACKGROUND"
	eventScreenOn        = "SCREEN_INTERACTIVE"
	eventScreenOff       = "SCREEN_NON_INTERACTIVE"
	eventKeyguardShown   = "KEYGUARD_SHOWN"
	eventKeyguardHidden  = "KEYGUARD_HIDDEN"
	eventDeviceShutdown  = "DEVICE_SHUTDOWN"
)

type Totals struct {
	Milliseconds int64
	Launches     int
	LastUsed     int64
}

func isResume(t string) bool {
	return t == eventActivityResumed || t == eventMoveForeground
}

func isLeave(t string) bool {
	return t == eventActivityPaused || t == eventActivityStopped || t == eventMoveBackground
}

func closesEverything(t string) bool {
	return t == eventScreenOff || t == eventKeyguardShown || t == eventDeviceShutdown
}

func Sorted(events []Event) []Event {
	out := make([]Event, len(events))
	copy(out, events)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time < out[j].Time })
	return out
}

func Sessionize(events []Event, clampStart, clampEnd int64) map[string]*Totals {
	totals := make(map[string]*Totals)
	open := make(map[string]int64)

	add := func(pkg string, start, end int64) {
		s, e := start, end
		if s < clampStart {
			s = clampStart
		}
		if e > clampEnd {
			e = clampEnd
		}
		if e <= s {
			return
		}
		acc := totals[pkg]
		if acc == nil {
			acc = &Totals{}
			totals[pkg] = acc
		}
		acc.Milliseconds += e - s
	}

	touch := func(pkg string, at int64) *Totals {
		acc := totals[pkg]
		if acc == nil {
			acc = &Totals{}
			totals[pkg] = acc
		}
		if at > acc.LastUsed {
			acc.LastUsed = at
		}
		return acc
	}

	for _, ev := range Sorted(events) {
		if ev.Time > clampEnd {
			break
		}
		switch {
		case isResume(ev.Type) && ev.Package != "":
			for pkg, start := range open {
				if pkg != ev.Package {
					add(pkg, start, ev.Time)
					delete(open, pkg)
				}
			}
			if start, ok := open[ev.Package]; ok {
				add(ev.Package, start, ev.Time)
			}
			open[ev.Package] = ev.Time
			acc := touch(ev.Package, ev.Time)
			if ev.Time >= clampStart && ev.Time <= clampEnd {
				acc.Launches++
			}
		case isLeave(ev.Type) && ev.Package != "":
			if start, ok := open[ev.Package]; ok {
				add(ev.Package, start, ev.Time)
				delete(open, ev.Package)
			}
			touch(ev.Package, ev.Time)
		case closesEverything(ev.Type):
			for pkg, start := range open {
				add(pkg, start, ev.Time)
				delete(open, pkg)
			}
		}
	}

	for pkg, start := range open {
		add(pkg, start, clampEnd)
	}

	for pkg, acc := range totals {
		if acc.Milliseconds == 0 && acc.Launches == 0 {
			delete(totals, pkg)
		}
	}
	return totals
}

func Current(events []Event, at int64) (string, int64) {
	var pkg string
	var since int64
	for _, ev := range Sorted(events) {
		if ev.Time > at {
			break
		}
		switch {
		case isResume(ev.Type) && ev.Package != "":
			if pkg != ev.Package {
				pkg, since = ev.Package, ev.Time
			}
		case isLeave(ev.Type) && ev.Package == pkg && pkg != "":
			pkg, since = "", 0
		case closesEverything(ev.Type):
			pkg, since = "", 0
		}
	}
	return pkg, since
}

func Unlocks(events []Event, from, to int64) int {
	count := 0
	for _, ev := range events {
		if ev.Type == eventKeyguardHidden && ev.Time >= from && ev.Time <= to {
			count++
		}
	}
	return count
}

func ScreenOn(events []Event, at int64) bool {
	on := false
	for _, ev := range Sorted(events) {
		if ev.Time > at {
			break
		}
		switch ev.Type {
		case eventScreenOn:
			on = true
		case eventScreenOff, eventDeviceShutdown:
			on = false
		}
	}
	return on
}
