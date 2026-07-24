package usage

type AppUsage struct {
	Package           string `json:"package"`
	Label             string `json:"label"`
	Category          string `json:"category"`
	ForegroundSeconds int64  `json:"foreground_seconds"`
	LaunchCount       int    `json:"launch_count"`
	LastUsed          string `json:"last_used,omitempty"`
}

type TodayStats struct {
	Date                   string     `json:"date"`
	TotalForegroundSeconds int64      `json:"total_foreground_seconds"`
	Unlocks                int        `json:"unlocks"`
	Apps                   []AppUsage `json:"apps"`
}

type Snapshot struct {
	DeviceID           string     `json:"device_id"`
	ChatID             int64      `json:"chat_id"`
	WindowStart        string     `json:"window_start"`
	WindowEnd          string     `json:"window_end"`
	ForegroundApp      string     `json:"foreground_app,omitempty"`
	ForegroundPackage  string     `json:"foreground_package,omitempty"`
	ForegroundCategory string     `json:"foreground_category,omitempty"`
	ForegroundSince    string     `json:"foreground_since,omitempty"`
	ScreenOn           bool       `json:"screen_on"`
	Apps               []AppUsage `json:"apps"`
	DistractingSeconds int64      `json:"distracting_seconds"`
	Today              TodayStats `json:"today"`
	ChangedSinceLast   bool       `json:"changed_since_last"`
	Seq                int64      `json:"seq"`
}

type Event struct {
	Time    int64
	Type    string
	Package string
}
