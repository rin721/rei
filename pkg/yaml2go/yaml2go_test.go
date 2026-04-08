package yaml2go

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	source := []byte("server:\n  host: 127.0.0.1\n  port: 9999\n")

	code, err := Convert("config", source)
	if err != nil {
		t.Fatalf("Convert() returned error: %v", err)
	}

	if !strings.Contains(code, "type Config struct") {
		t.Fatalf("generated code = %q, want root struct", code)
	}
	if !strings.Contains(code, "type ConfigServer struct") {
		t.Fatalf("generated code = %q, want nested struct", code)
	}
	if !strings.Contains(code, "`yaml:\"host\"`") {
		t.Fatalf("generated code = %q, want yaml tag", code)
	}
}
