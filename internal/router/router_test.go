package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rin721/rei/internal/handler"
	"github.com/rin721/rei/internal/middleware"
	"github.com/rin721/rei/internal/models"
	"github.com/rin721/rei/internal/repository"
	"github.com/rin721/rei/internal/service"
	authservice "github.com/rin721/rei/internal/service/auth"
	rbacservice "github.com/rin721/rei/internal/service/rbac"
	sampleservice "github.com/rin721/rei/internal/service/sample"
	userservice "github.com/rin721/rei/internal/service/user"
	pkgcrypto "github.com/rin721/rei/pkg/crypto"
	pkgdatabase "github.com/rin721/rei/pkg/database"
	pkgdbtx "github.com/rin721/rei/pkg/dbtx"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	"github.com/rin721/rei/types/constants"
	typesuser "github.com/rin721/rei/types/user"
)

func TestRouterSetupHealth(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	engine := New().Setup(middleware.MiddlewareConfig{
		AppName: "go-scaffold2",
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Header().Get(constants.HeaderTraceID) == "" {
		t.Fatal("health response missing trace header")
	}
}

func TestRouterProtectedRBACRoute(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "zh-CN.yaml"), []byte("message: 你好\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "en-US.yaml"), []byte("message: hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	i18nManager, err := pkgi18n.New(pkgi18n.Config{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
		LocaleDir:      dir,
	})
	if err != nil {
		t.Fatalf("pkgi18n.New() returned error: %v", err)
	}

	jwtManager, err := pkgjwt.New(pkgjwt.Config{
		Secret: "router-secret",
		Issuer: "router-test",
	})
	if err != nil {
		t.Fatalf("pkgjwt.New() returned error: %v", err)
	}

	cryptoService, err := pkgcrypto.New(pkgcrypto.Config{})
	if err != nil {
		t.Fatalf("pkgcrypto.New() returned error: %v", err)
	}

	rbacManager, err := pkgrbac.New(pkgrbac.Config{})
	if err != nil {
		t.Fatalf("pkgrbac.New() returned error: %v", err)
	}

	store, err := pkgdatabase.New(pkgdatabase.Config{
		Driver:          "sqlite",
		DSN:             filepath.Join(dir, "router.db"),
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
	})
	if err != nil {
		t.Fatalf("pkgdatabase.New() returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})

	dbtxManager, err := pkgdbtx.New(store.DB())
	if err != nil {
		t.Fatalf("pkgdbtx.New() returned error: %v", err)
	}
	if err := store.DB().AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{}, &models.Policy{}, &models.Sample{}); err != nil {
		t.Fatalf("AutoMigrate() returned error: %v", err)
	}

	repos := repository.NewSet(store.DB(), dbtxManager)
	if err := repos.Roles.Ensure(context.Background(), &models.Role{
		BaseModel:   models.BaseModel{ID: "role-admin"},
		Name:        service.DefaultRoleAdmin,
		Description: "system administrator",
	}); err != nil {
		t.Fatalf("Roles.Ensure(admin) returned error: %v", err)
	}
	if err := repos.Roles.Ensure(context.Background(), &models.Role{
		BaseModel:   models.BaseModel{ID: "role-user"},
		Name:        service.DefaultRoleUser,
		Description: "registered user",
	}); err != nil {
		t.Fatalf("Roles.Ensure(user) returned error: %v", err)
	}
	if err := repos.Policies.Add(context.Background(), &models.Policy{
		BaseModel: models.BaseModel{ID: "policy-1"},
		Subject:   service.DefaultRoleAdmin,
		Object:    RouteRBACCheck,
		Action:    "get",
	}); err != nil {
		t.Fatalf("Policies.Add(RouteRBACCheck) returned error: %v", err)
	}

	rbacService, err := rbacservice.New(rbacservice.Dependencies{
		Users:       repos.Users,
		Roles:       repos.Roles,
		UserRoles:   repos.UserRoles,
		Policies:    repos.Policies,
		IDProvider:  &testIDProvider{next: 100},
		Tx:          dbtxManager,
		RoleManager: rbacManager,
	})
	if err != nil {
		t.Fatalf("rbacservice.New() returned error: %v", err)
	}
	if err := rbacService.LoadFromStore(context.Background()); err != nil {
		t.Fatalf("LoadFromStore() returned error: %v", err)
	}

	authService, err := authservice.New(authservice.Dependencies{
		Users:           repos.Users,
		Roles:           repos.Roles,
		UserRoles:       repos.UserRoles,
		IDProvider:      &testIDProvider{next: 1000},
		Password:        cryptoService,
		Tokens:          jwtManager,
		Cache:           newTestCache(),
		Tx:              dbtxManager,
		RoleManager:     rbacManager,
		RefreshTokenTTL: 72 * time.Hour,
	})
	if err != nil {
		t.Fatalf("authservice.New() returned error: %v", err)
	}

	userService, err := userservice.New(userservice.Dependencies{
		Users:     repos.Users,
		UserRoles: repos.UserRoles,
	})
	if err != nil {
		t.Fatalf("userservice.New() returned error: %v", err)
	}

	sampleService, err := sampleservice.New(sampleservice.Dependencies{
		Samples: repos.Samples,
	})
	if err != nil {
		t.Fatalf("sampleservice.New() returned error: %v", err)
	}

	authResponse, err := authService.Register(context.Background(), typesuser.RegisterRequest{
		Username: "admin",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	engine := New(handler.NewBundle(authService, userService, rbacService, sampleService)).Setup(middleware.MiddlewareConfig{
		AppName: "go-scaffold2",
		I18n:    i18nManager,
		JWT:     jwtManager,
		RBAC:    rbacManager,
		CORS: middleware.CORSConfig{
			Enabled:      true,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Authorization", "Content-Type", "X-Trace-ID"},
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/rbac/check", nil)
	request.Header.Set("Authorization", "Bearer "+authResponse.Tokens.AccessToken)
	request.Header.Set(constants.HeaderAcceptLanguage, "en-US")
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Language"); got != "en-US" {
		t.Fatalf("Content-Language = %q, want %q", got, "en-US")
	}
}

type testIDProvider struct {
	next int64
}

func (p *testIDProvider) NextID() (int64, error) {
	p.next++
	return p.next, nil
}

type testCache struct {
	values map[string]any
}

func newTestCache() *testCache {
	return &testCache{
		values: make(map[string]any),
	}
}

func (c *testCache) Get(_ context.Context, key string) (any, bool) {
	value, ok := c.values[key]
	return value, ok
}

func (c *testCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	c.values[key] = value
	return nil
}

func (c *testCache) Delete(_ context.Context, key string) error {
	delete(c.values, key)
	return nil
}
