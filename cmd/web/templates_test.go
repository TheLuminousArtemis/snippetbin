package main

import (
	"testing"
	"time"

	"github.com/theluminousartemis/snippetbin/internal/assert"
)

func TestHumanDate(t *testing.T) {
	// tm := time.Date(2025, 6, 6, 12, 00, 0, 0, time.UTC)
	// hd := humanDate(tm)

	// if hd != "06 Jun 2025 at 12:00" {
	// 	t.Errorf("got %q; want %q;", hd, "06 Jun 2025 at 12:00")
	// }

	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2025, 6, 6, 12, 00, 0, 0, time.UTC),
			want: "06 Jun 2025 at 12:00",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2025, 6, 6, 12, 00, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "06 Jun 2025 at 12:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)
			assert.Equal(t, hd, tt.want)
		})
	}

}
