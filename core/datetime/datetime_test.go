package datetime

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "current time",
			time: time.Now(),
			want: time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			name: "specific date and time",
			time: time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
			want: "2022-01-01 12:00:00",
		},
		{
			name: "zero time",
			time: time.Time{},
			want: "0001-01-01 00:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTime(tt.time); got != tt.want {
				t.Errorf("FormatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormaRelativeTime(t *testing.T) {
	now := time.Now()

	// Test FormaRelativeTime with time difference less than 1 hour
	published := now.Add(-30 * time.Minute)
	if FormatRelativeTimeCN(published) != "30分钟前" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 30分钟前", published, FormatRelativeTimeCN(published))
	}
	if FormatRelativeTimeTW(published) != "30分鐘前" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 30分鐘前", published, FormatRelativeTimeTW(published))
	}
	if FormatRelativeTime(published) != "30m ago" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 30m ago", published, FormatRelativeTime(published))
	}
	published = now.Add(-1 * time.Minute)
	if FormatRelativeTime(published) != "1m ago" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 1m ago", published, FormatRelativeTime(published))
	}
	published = now.Add(-30 * time.Second)
	if FormatRelativeTimeCN(published) != "刚刚" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 刚刚", published, FormatRelativeTimeCN(published))
	}
	if FormatRelativeTime(published) != "Just now" {
		t.Errorf("FormaRelativeTime(%v) = %v, want Just now", published, FormatRelativeTime(published))
	}
	// Test FormaRelativeTime with time difference between 1 hour and 1 day
	published = now.Add(-2 * time.Hour)
	if FormatRelativeTimeCN(published) != "2小时前" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 2小时前", published, FormatRelativeTimeCN(published))
	}
	if FormatRelativeTime(published) != "2h ago" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 2h ago", published, FormatRelativeTime(published))
	}
	published = now.Add(-1 * time.Hour)
	if FormatRelativeTimeWithCurrentTime(now, published) != "1h ago" {
		t.Errorf("FormaRelativeTimeWithCurrentTime(%v, %v) = %v, want 1h ago", now, published, FormatRelativeTimeWithCurrentTime(now, published))
	}
	// Test FormaRelativeTime with time difference between 1 day and 3 days
	published = now.Add(-48 * time.Hour)
	if FormatRelativeTimeCN(published) != "2天前" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 2天前", published, FormatRelativeTimeCN(published))
	}
	if FormatRelativeTime(published) != "2d ago" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 2d ago", published, FormatRelativeTime(published))
	}
	published = now.Add(-24 * time.Hour)
	if FormatRelativeTime(published) != "1d ago" {
		t.Errorf("FormaRelativeTime(%v) = %v, want 1d ago", published, FormatRelativeTime(published))
	}

	// Test FormaRelativeTime with time difference between 3 days and 1 year
	published = now.Add(-73 * time.Hour)
	if now.Year() == published.Year() {

		if FormatRelativeTimeCN(published) != published.Format("01-02") {
			t.Errorf("FormaRelativeTimeCN(%v) = %s, want 1月前", published, FormatRelativeTimeCN(published))
		}
		if FormatRelativeTime(published) != published.Format("01-02") {
			t.Errorf("FormaRelativeTime(%v) = %s, want 1月前", published, FormatRelativeTime(published))
		}
	} else {
		if FormatRelativeTimeCN(published) != published.Format("2006-01-02") {
			t.Errorf("FormaRelativeTimeCN(%v) = %s, want 今年前", published, FormatRelativeTimeCN(published))
		}
		if FormatRelativeTime(published) != published.Format("2006-01-02") {
			t.Errorf("FormaRelativeTime(%v) = %s, want 今年前", published, FormatRelativeTime(published))
		}
	}

	// Test FormaRelativeTime with time difference greater than 1 year
	published = now.Add(-365 * 24 * time.Hour)
	if FormatRelativeTimeCN(published) != published.Format("2006-01-02") {
		t.Errorf("FormaRelativeTime(%v) = %s, want 今年前", published, FormatRelativeTimeCN(published))
	}
	if FormatRelativeTime(published) != published.Format("2006-01-02") {
		t.Errorf("FormaRelativeTime(%v) = %s, want 今年前", published, FormatRelativeTime(published))
	}
}

func TestFormaRelativeTimeWithCurrentTimeCN(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		now       time.Time
		published time.Time
		want      string
	}{
		{
			name:      "less than 1 hour",
			published: now.Add(-30 * time.Minute),
			now:       now,
			want:      "30分钟前",
		},
		{
			name:      "1 hour to less than 1 day",
			published: now.Add(-2 * time.Hour),
			now:       now,
			want:      "2小时前",
		},
		{
			name:      "1 day to less than 3 days",
			published: now.Add(-48 * time.Hour),
			now:       now,
			want:      "2天前",
		},

		{
			name:      "different year",
			published: now.Add(-365 * 24 * time.Hour),
			now:       now,
			want:      now.Add(-365 * 24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRelativeTimeWithCurrentTimeCN(tt.now, tt.published); got != tt.want {
				t.Errorf("FormaRelativeTimeWithCurrentTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormaRelativeTimeWithCurrentTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		now       time.Time
		published time.Time
		want      string
	}{
		{
			name:      "less than 1 hour",
			published: now.Add(-30 * time.Minute),
			now:       now,
			want:      "30m ago",
		},
		{
			name:      "1 hour to less than 1 day",
			published: now.Add(-2 * time.Hour),
			now:       now,
			want:      "2h ago",
		},
		{
			name:      "1 day to less than 3 days",
			published: now.Add(-48 * time.Hour),
			now:       now,
			want:      "2d ago",
		},
		{
			name:      "different year",
			published: now.Add(-365 * 24 * time.Hour),
			now:       now,
			want:      now.Add(-365 * 24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRelativeTimeWithCurrentTime(tt.now, tt.published); got != tt.want {
				t.Errorf("FormaRelativeTimeWithCurrentTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
