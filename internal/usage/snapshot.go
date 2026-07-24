package usage

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type AppMeta struct {
	Label    string
	Category string
}

type Params struct {
	DeviceID      string
	ChatID        int64
	WindowStart   time.Time
	WindowEnd     time.Time
	Seq           int64
	LastSignature string
	Location      *time.Location
	AppInfo       map[string]AppMeta
}

func (p Params) label(pkg string) string {
	if m, ok := p.AppInfo[pkg]; ok && m.Label != "" {
		return m.Label
	}
	return LabelOf(pkg)
}

func (p Params) category(pkg string) Category {
	if m, ok := p.AppInfo[pkg]; ok && m.Category != "" {
		return Category{ID: m.Category, Distracting: distractingID(m.Category)}
	}
	return CategoryOf(pkg)
}

type Result struct {
	Snapshot  Snapshot
	Signature string
}

func Build(events []Event, p Params) Result {
	loc := p.Location
	if loc == nil {
		loc = time.Local
	}

	startMs := p.WindowStart.UnixMilli()
	endMs := p.WindowEnd.UnixMilli()
	dayStart := startOfDay(p.WindowEnd, loc)
	dayStartMs := dayStart.UnixMilli()

	windowApps := totalsToApps(Sessionize(events, startMs, endMs), p)
	todayApps := totalsToApps(Sessionize(events, dayStartMs, endMs), p)

	var todayTotal int64
	for _, a := range todayApps {
		todayTotal += a.ForegroundSeconds
	}

	var distracting int64
	for _, a := range windowApps {
		if p.category(a.Package).Distracting {
			distracting += a.ForegroundSeconds
		}
	}

	currentPkg, since := Current(events, endMs)

	snapshot := Snapshot{
		DeviceID:           p.DeviceID,
		ChatID:             p.ChatID,
		WindowStart:        formatMillis(startMs),
		WindowEnd:          formatMillis(endMs),
		ScreenOn:           ScreenOn(events, endMs),
		Apps:               windowApps,
		DistractingSeconds: distracting,
		Today: TodayStats{
			Date:                   dayStart.Format("2006-01-02"),
			TotalForegroundSeconds: todayTotal,
			Unlocks:                Unlocks(events, dayStartMs, endMs),
			Apps:                   todayApps,
		},
		Seq: p.Seq,
	}

	if currentPkg != "" {
		snapshot.ForegroundPackage = currentPkg
		snapshot.ForegroundApp = p.label(currentPkg)
		snapshot.ForegroundCategory = p.category(currentPkg).ID
		if since > 0 {
			snapshot.ForegroundSince = formatMillis(since)
		}
	}

	signature := Signature(windowApps, currentPkg)
	snapshot.ChangedSinceLast = signature != p.LastSignature

	return Result{Snapshot: snapshot, Signature: signature}
}

func Signature(apps []AppUsage, current string) string {
	parts := make([]string, 0, len(apps)+1)
	for _, a := range apps {
		parts = append(parts, fmt.Sprintf("%s:%d", a.Package, a.ForegroundSeconds))
	}
	if current == "" {
		current = "-"
	}
	return strings.Join(parts, "|") + "#" + current
}

func totalsToApps(totals map[string]*Totals, p Params) []AppUsage {
	apps := make([]AppUsage, 0, len(totals))
	for pkg, acc := range totals {
		seconds := acc.Milliseconds / 1000
		if seconds == 0 && acc.Launches == 0 {
			continue
		}
		usage := AppUsage{
			Package:           pkg,
			Label:             p.label(pkg),
			Category:          p.category(pkg).ID,
			ForegroundSeconds: seconds,
			LaunchCount:       acc.Launches,
		}
		if acc.LastUsed > 0 {
			usage.LastUsed = formatMillis(acc.LastUsed)
		}
		apps = append(apps, usage)
	}
	sort.SliceStable(apps, func(i, j int) bool {
		if apps[i].ForegroundSeconds != apps[j].ForegroundSeconds {
			return apps[i].ForegroundSeconds > apps[j].ForegroundSeconds
		}
		return apps[i].Package < apps[j].Package
	})
	return apps
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
}

func formatMillis(ms int64) string {
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}
