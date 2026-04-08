package result

import (
	"testing"
	"time"

	apperrors "github.com/rei0721/go-scaffold2/types/errors"
)

func TestSuccess(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	got := Success(map[string]string{"status": "ok"}, "trace-1", now)

	if got.Code != apperrors.CodeOK {
		t.Fatalf("Code = %d, want %d", got.Code, apperrors.CodeOK)
	}

	if got.Message != "success" {
		t.Fatalf("Message = %q, want %q", got.Message, "success")
	}

	if got.ServerTime != now.Unix() {
		t.Fatalf("ServerTime = %d, want %d", got.ServerTime, now.Unix())
	}
}

func TestFailure(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000001, 0).UTC()
	got := Failure(apperrors.BadRequest("invalid payload"), "trace-2", now)

	if got.Code != apperrors.CodeBadRequest {
		t.Fatalf("Code = %d, want %d", got.Code, apperrors.CodeBadRequest)
	}

	if got.Message != "invalid payload" {
		t.Fatalf("Message = %q, want %q", got.Message, "invalid payload")
	}

	if got.Data != nil {
		t.Fatalf("Data = %#v, want nil", got.Data)
	}
}
