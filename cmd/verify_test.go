package cmd

import (
	"testing"
	"time"
)

func TestParseExpiryWindow(t *testing.T) {
	tests := []struct {
		input   string
		want    time.Duration
		wantErr bool
	}{
		{"", 0, false},
		{"30d", 30 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"30", 30 * 24 * time.Hour, false},
		{"720h", 720 * time.Hour, false},
		{"90m", 90 * time.Minute, false},
		{" 30d ", 30 * 24 * time.Hour, false},
		{"-30d", 0, true},
		{"-5h", 0, true},
		{"abc", 0, true},
		{"d", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseExpiryWindow(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseExpiryWindow(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseExpiryWindow(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
