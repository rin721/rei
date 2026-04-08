package errors

import (
	stderrors "errors"
	"testing"
)

func TestCodeOfWrappedError(t *testing.T) {
	t.Parallel()

	err := Wrap(CodeForbidden, "permission denied", stderrors.New("rbac denied"))

	if got := CodeOf(err); got != CodeForbidden {
		t.Fatalf("CodeOf() = %d, want %d", got, CodeForbidden)
	}
}

func TestMessageOfPlainError(t *testing.T) {
	t.Parallel()

	err := stderrors.New("boom")

	if got := MessageOf(err); got != "boom" {
		t.Fatalf("MessageOf() = %q, want %q", got, "boom")
	}
}

func TestNewUsesDefaultMessageWhenEmpty(t *testing.T) {
	t.Parallel()

	err := New(CodeNotFound, "")

	if err.Message != "not found" {
		t.Fatalf("Message = %q, want %q", err.Message, "not found")
	}
}
