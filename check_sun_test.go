package main

import (
	"testing"
	"time"
)

// Today (2026-04-05) at location 53.19N 19.76E (Poland, CEST = UTC+2):
//   sunrise ~06:05:28 local  (user's weather app showed 06:03)
//   sunset  ~19:21:35 local  (user's weather app showed 19:22)

var warsaw, _ = time.LoadLocation("Europe/Warsaw")

func localTime(hour, min int) time.Time {
	return time.Date(2026, 4, 5, hour, min, 0, 0, warsaw)
}

func TestCheckIfSunIsDownAt(t *testing.T) {
	tests := []struct {
		name                    string
		now                     time.Time
		delayAfterSunset        int
		accelerateBeforeSunrise int
		wantDark                bool
	}{
		// --- no modifiers ---
		{
			name:     "deep night 03:00 → dark",
			now:      localTime(3, 0),
			wantDark: true,
		},
		{
			name:     "3 min before sunrise 06:00 → dark",
			now:      localTime(6, 0),
			wantDark: true,
		},
		{
			name:     "5 min after sunrise 06:08 → light",
			now:      localTime(6, 8),
			wantDark: false,
		},
		{
			name:     "midday 12:00 → light",
			now:      localTime(12, 0),
			wantDark: false,
		},
		{
			name:     "2 min before sunset 19:20 → light",
			now:      localTime(19, 20),
			wantDark: false,
		},
		{
			name:     "5 min after sunset 19:27 → dark",
			now:      localTime(19, 27),
			wantDark: true,
		},
		{
			name:     "late night 23:00 → dark",
			now:      localTime(23, 0),
			wantDark: true,
		},

		// --- accelerateBeforeSunrise: treat as daytime N min before actual sunrise ---
		{
			name:                    "before=15: 05:51 + 15min = 06:06 > 06:05:28 → light",
			now:                     localTime(5, 51),
			accelerateBeforeSunrise: 15,
			wantDark:                false,
		},
		{
			name:                    "before=15: 05:49 + 15min = 06:04 < 06:05:28 → dark",
			now:                     localTime(5, 49),
			accelerateBeforeSunrise: 15,
			wantDark:                true,
		},
		{
			name:                    "before=5: 06:01 + 5min = 06:06 > 06:05:28 → light",
			now:                     localTime(6, 1),
			accelerateBeforeSunrise: 5,
			wantDark:                false,
		},
		{
			name:                    "before=5: 05:59 + 5min = 06:04 < 06:05:28 → dark",
			now:                     localTime(5, 59),
			accelerateBeforeSunrise: 5,
			wantDark:                true,
		},

		// --- delayAfterSunset: treat as daytime N min after actual sunset ---
		{
			name:             "after=30: 19:40 - 30min = 19:10 ≤ 19:22 → light",
			now:              localTime(19, 40),
			delayAfterSunset: 30,
			wantDark:         false,
		},
		{
			name:             "after=30: 19:55 - 30min = 19:25 > 19:22 → dark",
			now:              localTime(19, 55),
			delayAfterSunset: 30,
			wantDark:         true,
		},
		{
			name:             "after=10: 19:30 - 10min = 19:20 ≤ 19:22 → light",
			now:              localTime(19, 30),
			delayAfterSunset: 10,
			wantDark:         false,
		},
		{
			name:             "after=10: 19:35 - 10min = 19:25 > 19:22 → dark",
			now:              localTime(19, 35),
			delayAfterSunset: 10,
			wantDark:         true,
		},

		// --- both modifiers together ---
		{
			name:                    "before=15 after=30: midday 12:00 → light",
			now:                     localTime(12, 0),
			delayAfterSunset:        30,
			accelerateBeforeSunrise: 15,
			wantDark:                false,
		},
		{
			name:                    "before=15 after=30: 05:51 → light (before window covers it)",
			now:                     localTime(5, 51),
			delayAfterSunset:        30,
			accelerateBeforeSunrise: 15,
			wantDark:                false,
		},
		{
			name:                    "before=15 after=30: 05:49 → dark (before window not enough)",
			now:                     localTime(5, 49),
			delayAfterSunset:        30,
			accelerateBeforeSunrise: 15,
			wantDark:                true,
		},
		{
			name:                    "before=15 after=30: 19:50 → light (after window covers it)",
			now:                     localTime(19, 50),
			delayAfterSunset:        30,
			accelerateBeforeSunrise: 15,
			wantDark:                false,
		},
		{
			name:                    "before=15 after=30: 19:55 → dark (after window not enough)",
			now:                     localTime(19, 55),
			delayAfterSunset:        30,
			accelerateBeforeSunrise: 15,
			wantDark:                true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkIfSunIsDownAt(tc.now, tc.delayAfterSunset, tc.accelerateBeforeSunrise)
			if got != tc.wantDark {
				t.Errorf("checkIfSunIsDownAt(%s, after=%d, before=%d) = %v, want %v",
					tc.now.Format("15:04 MST"), tc.delayAfterSunset, tc.accelerateBeforeSunrise, got, tc.wantDark)
			}
		})
	}
}
