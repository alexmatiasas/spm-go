package logging

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name   string
		want   slog.Level
		wantOK bool
	}{
		{"debug", slog.LevelDebug, true},
		{"DEBUG", slog.LevelDebug, true},
		{"info", slog.LevelInfo, true},
		{"Info", slog.LevelInfo, true},
		{"warn", slog.LevelWarn, true},
		{"warning", slog.LevelWarn, true},
		{"error", slog.LevelError, true},
		{"", slog.LevelInfo, true},
		{"banana", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseLevel(tt.name)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("ParseLevel(%q) = (%v, %v), want (%v, %v)", tt.name, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestSetupRoutesByLevel(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := Setup(buf, "debug")
	logger.Info("hello info")
	if !strings.Contains(buf.String(), "hello info") {
		t.Errorf("info not written at debug level, got %q", buf.String())
	}

	buf.Reset()
	quiet := Setup(buf, "error")
	quiet.Info("noisy")
	if strings.Contains(buf.String(), "noisy") {
		t.Errorf("info should be silenced at error level, got %q", buf.String())
	}
	quiet.Error("boom")
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("error not written at error level, got %q", buf.String())
	}
}

func TestSetupUnknownLevelFallsBackToInfo(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := Setup(buf, "banana")
	logger.Info("still visible")
	if !strings.Contains(buf.String(), "still visible") {
		t.Errorf("unknown level should fall back to info, got %q", buf.String())
	}
}

func TestSetupReplacesDefaultLogger(t *testing.T) {
	t.Cleanup(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	})

	buf := &bytes.Buffer{}
	Setup(buf, "debug")

	slog.Debug("via default logger")
	if !strings.Contains(buf.String(), "via default logger") {
		t.Errorf("slog.Default() not routed to setup writer, got %q", buf.String())
	}
}
