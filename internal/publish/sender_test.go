package publish

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"http://10.0.2.2:8099", "http://10.0.2.2:8099"},
		{"Http://host:8083", "http://host:8083"},
		{"HTTPS://Host.Example/", "https://Host.Example"},
		{"  https://host:8083/  ", "https://host:8083"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := normalizeURL(tc.in); got != tc.want {
			t.Errorf("normalizeURL(%q) = %q, ожидалось %q", tc.in, got, tc.want)
		}
	}
}
