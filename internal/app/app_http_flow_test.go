package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authEnvelopeData struct {
	User struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Roles    []string `json:"roles"`
	} `json:"user"`
	Tokens struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	} `json:"tokens"`
}

func TestAppHTTPFlowAuthAndUser(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)
	prepareTestSchema(t, configPath)
	application, err := New(Options{
		Mode:       ModeServer,
		ConfigPath: configPath,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = application.Shutdown(context.Background())
	})

	if err := application.bootstrapServer(context.Background()); err != nil {
		t.Fatalf("bootstrapServer() returned error: %v", err)
	}

	registerResponse := performJSONRequest(t, application.httpHandler(), http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username":    "alice",
		"password":    "Password123",
		"displayName": "Alice",
	})
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d", registerResponse.Code, http.StatusCreated)
	}

	registerBody := decodeEnvelope(t, registerResponse.Body.Bytes())
	if registerBody.Code != 0 {
		t.Fatalf("register envelope code = %d, want 0", registerBody.Code)
	}

	var authData authEnvelopeData
	if err := json.Unmarshal(registerBody.Data, &authData); err != nil {
		t.Fatalf("json.Unmarshal(register data) returned error: %v", err)
	}
	if authData.User.Username != "alice" {
		t.Fatalf("register username = %q, want %q", authData.User.Username, "alice")
	}
	if authData.Tokens.AccessToken == "" || authData.Tokens.RefreshToken == "" {
		t.Fatal("register tokens should not be empty")
	}

	profileResponse := performJSONRequest(t, application.httpHandler(), http.MethodGet, "/api/v1/users/me", nil, withBearer(authData.Tokens.AccessToken))
	if profileResponse.Code != http.StatusOK {
		t.Fatalf("profile status = %d, want %d", profileResponse.Code, http.StatusOK)
	}

	refreshResponse := performJSONRequest(t, application.httpHandler(), http.MethodPost, "/api/v1/auth/refresh", map[string]any{
		"refreshToken": authData.Tokens.RefreshToken,
	})
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, want %d", refreshResponse.Code, http.StatusOK)
	}

	changePasswordResponse := performJSONRequest(t, application.httpHandler(), http.MethodPost, "/api/v1/auth/change-password", map[string]any{
		"oldPassword": "Password123",
		"newPassword": "Password456",
	}, withBearer(authData.Tokens.AccessToken))
	if changePasswordResponse.Code != http.StatusOK {
		t.Fatalf("change password status = %d, want %d", changePasswordResponse.Code, http.StatusOK)
	}

	loginResponse := performJSONRequest(t, application.httpHandler(), http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "alice",
		"password": "Password456",
	})
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d", loginResponse.Code, http.StatusOK)
	}
}

type requestOption func(*http.Request)

func withBearer(token string) requestOption {
	return func(request *http.Request) {
		request.Header.Set("Authorization", "Bearer "+token)
	}
}

func performJSONRequest(t *testing.T, handler http.Handler, method, path string, body any, options ...requestOption) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal() returned error: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(request)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeEnvelope(t *testing.T, body []byte) envelope {
	t.Helper()

	var response envelope
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("json.Unmarshal(envelope) returned error: %v", err)
	}
	return response
}
