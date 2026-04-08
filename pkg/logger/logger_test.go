package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerRespectsLevel(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	log, err := New(Config{
		Level:  "info",
		Writer: &buffer,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	log.Debug("hidden")
	log.Info("visible")

	output := buffer.String()
	if strings.Contains(output, "hidden") {
		t.Fatalf("output = %q, should not contain debug log", output)
	}
	if !strings.Contains(output, "visible") {
		t.Fatalf("output = %q, should contain info log", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	log, err := New(Config{
		Level:  "debug",
		Writer: &buffer,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	log.With(map[string]any{"component": "auth"}).Info("started")

	output := buffer.String()
	if !strings.Contains(output, `component="auth"`) {
		t.Fatalf("output = %q, should contain structured field", output)
	}
}
