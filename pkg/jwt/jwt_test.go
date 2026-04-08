package jwt

import "testing"

func TestManagerGenerateValidateAndRefreshToken(t *testing.T) {
	t.Parallel()

	manager, err := New(Config{
		Issuer: "phase2",
		Secret: "secret-key",
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	refreshToken, err := manager.GenerateToken("user-1", TokenTypeRefresh, map[string]any{"scope": "basic"})
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	claims, err := manager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("ValidateToken() returned error: %v", err)
	}
	if claims.Subject != "user-1" {
		t.Fatalf("Subject = %q, want %q", claims.Subject, "user-1")
	}
	if claims.TokenType != string(TokenTypeRefresh) {
		t.Fatalf("TokenType = %q, want %q", claims.TokenType, TokenTypeRefresh)
	}

	pair, err := manager.RefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("RefreshToken() returned error: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("RefreshToken() returned empty token pair")
	}
}
