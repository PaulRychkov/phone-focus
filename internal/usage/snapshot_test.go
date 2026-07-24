package usage

import (
	"testing"
	"time"
)

func TestBuildUsesAppInfo(t *testing.T) {
	base := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Time: base.Add(-5 * time.Minute).UnixMilli(), Type: eventActivityResumed, Package: "com.supercell.clashroyale"},
		{Time: base.Add(-1 * time.Minute).UnixMilli(), Type: eventActivityPaused, Package: "com.supercell.clashroyale"},
	}

	p := Params{
		DeviceID:    "dev",
		WindowStart: base.Add(-10 * time.Minute),
		WindowEnd:   base,
		Location:    time.UTC,
		AppInfo: map[string]AppMeta{
			"com.supercell.clashroyale": {Label: "Clash Royale", Category: "games"},
		},
	}

	res := Build(events, p)
	if len(res.Snapshot.Apps) != 1 {
		t.Fatalf("приложений за окно %d, ожидалось 1", len(res.Snapshot.Apps))
	}
	app := res.Snapshot.Apps[0]
	if app.Label != "Clash Royale" {
		t.Errorf("название %q, ожидалось системное 'Clash Royale'", app.Label)
	}
	if app.Category != "games" {
		t.Errorf("категория %q, ожидалась системная 'games'", app.Category)
	}
	if res.Snapshot.DistractingSeconds < 200 {
		t.Errorf("игра должна попасть в отвлекающие, distracting=%d", res.Snapshot.DistractingSeconds)
	}
}

func TestBuildFallsBackWithoutAppInfo(t *testing.T) {
	base := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	events := []Event{
		{Time: base.Add(-3 * time.Minute).UnixMilli(), Type: eventActivityResumed, Package: "com.google.android.youtube"},
		{Time: base.Add(-1 * time.Minute).UnixMilli(), Type: eventActivityPaused, Package: "com.google.android.youtube"},
	}
	p := Params{WindowStart: base.Add(-10 * time.Minute), WindowEnd: base, Location: time.UTC}
	res := Build(events, p)
	if res.Snapshot.Apps[0].Category != "video" {
		t.Errorf("без метаданных должен сработать словарь: категория %q, ожидалась 'video'", res.Snapshot.Apps[0].Category)
	}
	if res.Snapshot.DistractingSeconds == 0 {
		t.Error("youtube из словаря должен считаться отвлекающим")
	}
}
