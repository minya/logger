package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rs/zerolog"
)

func capture(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	old := DefaultLogger
	DefaultLogger = zerolog.New(&buf)
	t.Cleanup(func() { DefaultLogger = old })
	return &buf
}

func message(t *testing.T, buf *bytes.Buffer) string {
	t.Helper()
	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log line %q: %v", buf.String(), err)
	}
	buf.Reset()
	s, _ := rec["message"].(string)
	return s
}

// F63: messages are printf-formatted; nothing is mangled or dropped.
func TestMessagesArePrintfFormatted(t *testing.T) {
	buf := capture(t)
	for _, tc := range []struct {
		log  func()
		want string
	}{
		{func() { Warn("torrent %s at %d%%", "x", 50) }, "torrent x at 50%"},
		// F132: vet requires %% even without arguments; it must print as %.
		{func() { Info("progress 100%% done") }, "progress 100% done"},
		{func() { Info("plain message") }, "plain message"},
		{func() { Error(errors.New("e"), "Got value %d", 42) }, "Got value 42"},
	} {
		tc.log()
		if got := message(t, buf); got != tc.want {
			t.Errorf("message = %q, want %q", got, tc.want)
		}
	}
}

// F63: an unknown level is reported instead of silently becoming info.
func TestParseLevel(t *testing.T) {
	for in, want := range map[string]zerolog.Level{
		"": zerolog.InfoLevel, "debug": zerolog.DebugLevel, "WARN": zerolog.WarnLevel,
		"warning": zerolog.WarnLevel, "trace": zerolog.TraceLevel, "disabled": zerolog.Disabled,
	} {
		if got, err := parseLevel(in); err != nil || got != want {
			t.Errorf("parseLevel(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	// F113: numbers are refused too; zerolog would take "8" as a custom
	// level above panic and log nothing at all.
	for _, in := range []string{"verbose", "8", "10", "100", "-5", "1"} {
		if got, err := parseLevel(in); err == nil || got != zerolog.InfoLevel {
			t.Errorf("parseLevel(%q) = %v, %v; want info and an error", in, got, err)
		}
	}
}
