package jwt

import (
	"errors"
	"fmt"
	"sync"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

// TokenType 表示 JWT 的业务类型。
type TokenType string

const (
	// TokenTypeAccess 表示访问令牌。
	TokenTypeAccess TokenType = "access"
	// TokenTypeRefresh 表示刷新令牌。
	TokenTypeRefresh TokenType = "refresh"
)

// Config 描述 JWT 配置。
type Config struct {
	Issuer     string
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// Claims 定义统一 JWT 声明结构。
type Claims struct {
	TokenType string         `json:"token_type"`
	Extra     map[string]any `json:"extra,omitempty"`
	gjwt.RegisteredClaims
}

// TokenPair 描述访问令牌和刷新令牌对。
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Manager 提供 JWT 生成、校验和刷新能力。
type Manager struct {
	mu  sync.RWMutex
	cfg Config
}

// New 创建一个新的 JWT 管理器。
func New(cfg Config) (*Manager, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{cfg: normalized}, nil
}

// GenerateToken 生成指定类型的 JWT。
func (m *Manager) GenerateToken(subject string, tokenType TokenType, extra map[string]any) (string, error) {
	if subject == "" {
		return "", errors.New("subject is required")
	}

	m.mu.RLock()
	cfg := m.cfg
	m.mu.RUnlock()

	ttl, err := ttlForType(cfg, tokenType)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	claims := Claims{
		TokenType: string(tokenType),
		Extra:     cloneMap(extra),
		RegisteredClaims: gjwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    cfg.Issuer,
			IssuedAt:  gjwt.NewNumericDate(now),
			ExpiresAt: gjwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ValidateToken 校验 JWT 并返回解析后的声明。
func (m *Manager) ValidateToken(token string) (*Claims, error) {
	m.mu.RLock()
	cfg := m.cfg
	m.mu.RUnlock()

	claims := &Claims{}
	parsed, err := gjwt.ParseWithClaims(token, claims, func(t *gjwt.Token) (any, error) {
		if _, ok := t.Method.(*gjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	}, gjwt.WithIssuer(cfg.Issuer))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken 使用刷新令牌生成新的访问令牌与刷新令牌。
func (m *Manager) RefreshToken(refreshToken string) (TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return TokenPair{}, err
	}
	if claims.TokenType != string(TokenTypeRefresh) {
		return TokenPair{}, errors.New("token is not a refresh token")
	}

	accessToken, err := m.GenerateToken(claims.Subject, TokenTypeAccess, claims.Extra)
	if err != nil {
		return TokenPair{}, err
	}
	newRefreshToken, err := m.GenerateToken(claims.Subject, TokenTypeRefresh, claims.Extra)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Reload 更新 JWT 配置。
func (m *Manager) Reload(cfg Config) error {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = normalized
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.Issuer == "" {
		cfg.Issuer = "go-scaffold2"
	}
	if cfg.Secret == "" {
		return Config{}, errors.New("jwt secret is required")
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = time.Hour
	}
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 72 * time.Hour
	}
	return cfg, nil
}

func ttlForType(cfg Config, tokenType TokenType) (time.Duration, error) {
	switch tokenType {
	case TokenTypeAccess:
		return cfg.AccessTTL, nil
	case TokenTypeRefresh:
		return cfg.RefreshTTL, nil
	default:
		return 0, fmt.Errorf("unsupported token type %q", tokenType)
	}
}

func cloneMap(source map[string]any) map[string]any {
	if len(source) == 0 {
		return nil
	}

	target := make(map[string]any, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}
