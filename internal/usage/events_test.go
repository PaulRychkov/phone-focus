package usage

import "testing"

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []Event
	}{
		{
			name: "обычные строки",
			raw:  "1700000000000\t1\tcom.a\n1700000005000\t2\tcom.a\n",
			want: []Event{
				{Time: 1700000000000, Type: eventActivityResumed, Package: "com.a"},
				{Time: 1700000005000, Type: eventActivityPaused, Package: "com.a"},
			},
		},
		{
			name: "событие без пакета",
			raw:  "1700000000000\t16\t\n",
			want: []Event{{Time: 1700000000000, Type: eventScreenOff}},
		},
		{
			name: "неизвестный код пропускается",
			raw:  "1700000000000\t99\tcom.a\n1700000001000\t1\tcom.b\n",
			want: []Event{{Time: 1700000001000, Type: eventActivityResumed, Package: "com.b"}},
		},
		{
			name: "битые строки пропускаются",
			raw:  "мусор\n\n1700000000000\n1700000002000\t23\tcom.c\n",
			want: []Event{{Time: 1700000002000, Type: eventActivityStopped, Package: "com.c"}},
		},
		{
			name: "пустой ввод",
			raw:  "",
			want: []Event{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseEvents(tc.raw)
			if len(got) != len(tc.want) {
				t.Fatalf("событий %d, ожидалось %d", len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("событие %d: %+v, ожидалось %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestCategoryOf(t *testing.T) {
	tests := []struct {
		pkg         string
		id          string
		distracting bool
	}{
		{"com.google.android.youtube", "video", true},
		{"org.telegram.messenger", "communication", false},
		{"com.example.supergame", "games", true},
		{"com.unknown.thing", "other", false},
	}
	for _, tc := range tests {
		t.Run(tc.pkg, func(t *testing.T) {
			got := CategoryOf(tc.pkg)
			if got.ID != tc.id || got.Distracting != tc.distracting {
				t.Errorf("получено %+v, ожидалось %s/%v", got, tc.id, tc.distracting)
			}
		})
	}
}
