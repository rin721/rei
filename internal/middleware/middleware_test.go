package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkglogger "github.com/rin721/rei/pkg/logger"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	"github.com/rin721/rei/types/constants"
)

func TestRecoveryReturnsEnvelope(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID(MiddlewareConfig{}), Recovery(MiddlewareConfig{}))
	engine.GET("/panic", func(_ *gin.Context) {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if traceID := recorder.Header().Get(constants.HeaderTraceID); traceID == "" {
		t.Fatal("response missing trace header")
	}
}

func TestAuthAndRBAC(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	jwtManager, err := pkgjwt.New(pkgjwt.Config{
		Secret: "phase4-secret",
		Issuer: "phase4",
	})
	if err != nil {
		t.Fatalf("pkgjwt.New() returned error: %v", err)
	}
	rbacManager, err := pkgrbac.New(pkgrbac.Config{})
	if err != nil {
		t.Fatalf("pkgrbac.New() returned error: %v", err)
	}
	if err := rbacManager.AddPolicy("admin", "/secure", "get"); err != nil {
		t.Fatalf("AddPolicy() returned error: %v", err)
	}
	if err := rbacManager.AssignRole("user-1", "admin"); err != nil {
		t.Fatalf("AssignRole() returned error: %v", err)
	}

	token, err := jwtManager.GenerateToken("user-1", pkgjwt.TokenTypeAccess, nil)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	engine := gin.New()
	engine.Use(TraceID(MiddlewareConfig{}))
	engine.GET("/secure",
		Auth(MiddlewareConfig{JWT: jwtManager}),
		RBAC(MiddlewareConfig{RBAC: rbacManager}, nil),
		func(c *gin.Context) {
			writeSuccess(c, http.StatusOK, gin.H{"ok": true})
		},
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/secure", nil)
	request.Header.Set(authorizationHeader, bearerPrefix+token)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestI18nAndLogger(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "zh-CN.yaml"), []byte("message: 你好\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "en-US.yaml"), []byte("message: hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	i18nManager, err := pkgi18n.New(&pkgi18n.Config{
		DefaultLanguage:    "zh-CN",
		SupportedLanguages: []string{"zh-CN", "en-US"},
		MessagesDir:        dir,
	})
	if err != nil {
		t.Fatalf("pkgi18n.New() returned error: %v", err)
	}

	logger := &testLogger{}

	engine := gin.New()
	engine.Use(I18n(MiddlewareConfig{I18n: i18nManager}), TraceID(MiddlewareConfig{}), Logger(MiddlewareConfig{Logger: logger}))
	engine.GET("/locale", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/locale", nil)
	request.Header.Set(constants.HeaderAcceptLanguage, "en-US")
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Content-Language"); got != "en-US" {
		t.Fatalf("Content-Language = %q, want %q", got, "en-US")
	}
	if len(logger.messages) == 0 {
		t.Fatal("logger did not write request log")
	}
}

type testLogger struct {
	messages []string
}

func (l *testLogger) Debug(string, ...interface{}) {}

func (l *testLogger) Info(msg string, _ ...interface{}) {
	l.messages = append(l.messages, msg)
}

func (l *testLogger) Warn(string, ...interface{}) {}

func (l *testLogger) Error(string, ...interface{}) {}

func (l *testLogger) Fatal(string, ...interface{}) {}

func (l *testLogger) With(...interface{}) pkglogger.Logger {
	return l
}

func (l *testLogger) Sync() error {
	return nil
}

func (l *testLogger) Reload(*pkglogger.Config) error {
	return nil
}
