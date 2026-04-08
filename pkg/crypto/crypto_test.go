package crypto

import "testing"

func TestServiceHashAndComparePassword(t *testing.T) {
	t.Parallel()

	service, err := New(Config{})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	hash, err := service.HashPassword("secret-123")
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}

	if err := service.ComparePassword(hash, "secret-123"); err != nil {
		t.Fatalf("ComparePassword() returned error: %v", err)
	}
}

func TestServiceReloadRejectsInvalidCost(t *testing.T) {
	t.Parallel()

	service, err := New(Config{})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := service.Reload(Config{Cost: -1}); err == nil {
		t.Fatal("Reload() returned nil error for invalid cost")
	}
}
