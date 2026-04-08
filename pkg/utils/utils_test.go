package utils

import "testing"

func TestIDGeneratorNextID(t *testing.T) {
	t.Parallel()

	generator, err := NewIDGenerator(1)
	if err != nil {
		t.Fatalf("NewIDGenerator() returned error: %v", err)
	}

	first, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}

	second, err := generator.NextID()
	if err != nil {
		t.Fatalf("NextID() returned error: %v", err)
	}

	if second <= first {
		t.Fatalf("second id = %d, want > %d", second, first)
	}
}

func TestGetFreePort(t *testing.T) {
	t.Parallel()

	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort() returned error: %v", err)
	}

	if port <= 0 {
		t.Fatalf("port = %d, want positive", port)
	}
}
