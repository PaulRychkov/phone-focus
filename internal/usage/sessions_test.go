package usage

import "testing"

const base int64 = 1700000000000

func TestSessionize(t *testing.T) {
	tests := []struct {
		name     string
		events   []Event
		start    int64
		end      int64
		seconds  map[string]int64
		launches map[string]int
	}{
		{
			name: "полная сессия внутри окна",
			events: []Event{
				{Time: base + 1000, Type: eventActivityResumed, Package: "a"},
				{Time: base + 61000, Type: eventActivityPaused, Package: "a"},
			},
			start:    base,
			end:      base + 120000,
			seconds:  map[string]int64{"a": 60},
			launches: map[string]int{"a": 1},
		},
		{
			name: "сессия началась до окна, считаем только попавшую часть",
			events: []Event{
				{Time: base - 60000, Type: eventActivityResumed, Package: "a"},
				{Time: base + 30000, Type: eventActivityPaused, Package: "a"},
			},
			start:    base,
			end:      base + 120000,
			seconds:  map[string]int64{"a": 30},
			launches: map[string]int{"a": 0},
		},
		{
			name: "незакрытая сессия досчитывается до конца окна",
			events: []Event{
				{Time: base + 10000, Type: eventActivityResumed, Package: "a"},
			},
			start:    base,
			end:      base + 70000,
			seconds:  map[string]int64{"a": 60},
			launches: map[string]int{"a": 1},
		},
		{
			name: "переключение закрывает предыдущее приложение",
			events: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 10000, Type: eventActivityResumed, Package: "b"},
				{Time: base + 30000, Type: eventActivityPaused, Package: "b"},
			},
			start:    base,
			end:      base + 30000,
			seconds:  map[string]int64{"a": 10, "b": 20},
			launches: map[string]int{"a": 1, "b": 1},
		},
		{
			name: "выключение экрана закрывает сессию",
			events: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 10000, Type: eventScreenOff},
			},
			start:   base,
			end:     base + 60000,
			seconds: map[string]int64{"a": 10},
		},
		{
			name: "ACTIVITY_STOPPED тоже закрывает сессию",
			events: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 5000, Type: eventActivityStopped, Package: "a"},
			},
			start:   base,
			end:     base + 60000,
			seconds: map[string]int64{"a": 5},
		},
		{
			name: "старые имена событий работают так же",
			events: []Event{
				{Time: base, Type: eventMoveForeground, Package: "a"},
				{Time: base + 20000, Type: eventMoveBackground, Package: "a"},
			},
			start:   base,
			end:     base + 60000,
			seconds: map[string]int64{"a": 20},
		},
		{
			name:    "события вне окна не учитываются",
			events:  []Event{{Time: base - 100000, Type: eventActivityResumed, Package: "a"}, {Time: base - 90000, Type: eventActivityPaused, Package: "a"}},
			start:   base,
			end:     base + 60000,
			seconds: map[string]int64{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Sessionize(tc.events, tc.start, tc.end)
			if len(got) != len(tc.seconds) {
				t.Fatalf("приложений: получено %d, ожидалось %d", len(got), len(tc.seconds))
			}
			for pkg, want := range tc.seconds {
				acc, ok := got[pkg]
				if !ok {
					t.Fatalf("нет данных по %s", pkg)
				}
				if acc.Milliseconds/1000 != want {
					t.Errorf("%s: секунд %d, ожидалось %d", pkg, acc.Milliseconds/1000, want)
				}
			}
			for pkg, want := range tc.launches {
				if got[pkg].Launches != want {
					t.Errorf("%s: запусков %d, ожидалось %d", pkg, got[pkg].Launches, want)
				}
			}
		})
	}
}

func TestCurrent(t *testing.T) {
	tests := []struct {
		name  string
		items []Event
		at    int64
		pkg   string
		since int64
	}{
		{
			name:  "открытое приложение",
			items: []Event{{Time: base, Type: eventActivityResumed, Package: "a"}},
			at:    base + 5000,
			pkg:   "a",
			since: base,
		},
		{
			name: "после паузы никого",
			items: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 1000, Type: eventActivityPaused, Package: "a"},
			},
			at:  base + 5000,
			pkg: "",
		},
		{
			name: "после переключения текущее последнее",
			items: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 1000, Type: eventActivityResumed, Package: "b"},
			},
			at:    base + 5000,
			pkg:   "b",
			since: base + 1000,
		},
		{
			name: "экран выключен - никого",
			items: []Event{
				{Time: base, Type: eventActivityResumed, Package: "a"},
				{Time: base + 1000, Type: eventScreenOff},
			},
			at:  base + 5000,
			pkg: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pkg, since := Current(tc.items, tc.at)
			if pkg != tc.pkg {
				t.Fatalf("пакет %q, ожидался %q", pkg, tc.pkg)
			}
			if tc.pkg != "" && since != tc.since {
				t.Errorf("since %d, ожидалось %d", since, tc.since)
			}
		})
	}
}

func TestUnlocksAndScreen(t *testing.T) {
	events := []Event{
		{Time: base + 1000, Type: eventKeyguardHidden},
		{Time: base + 2000, Type: eventScreenOn},
		{Time: base + 5000, Type: eventKeyguardHidden},
		{Time: base + 9000, Type: eventScreenOff},
		{Time: base + 20000, Type: eventKeyguardHidden},
	}

	if got := Unlocks(events, base, base+10000); got != 2 {
		t.Errorf("разблокировок %d, ожидалось 2", got)
	}
	if ScreenOn(events, base+9500) {
		t.Error("экран должен быть выключен")
	}
	if !ScreenOn(events, base+3000) {
		t.Error("экран должен быть включён")
	}
}
