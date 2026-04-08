package crypto

import (
	"errors"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Config 描述密码处理配置。
type Config struct {
	Cost int
}

// Service 提供密码哈希与校验能力。
type Service struct {
	mu   sync.RWMutex
	cost int
}

// New 创建一个新的密码服务。
func New(cfg Config) (*Service, error) {
	cost, err := normalizeCost(cfg.Cost)
	if err != nil {
		return nil, err
	}

	return &Service{cost: cost}, nil
}

// HashPassword 生成密码哈希。
func (s *Service) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password is required")
	}

	s.mu.RLock()
	cost := s.cost
	s.mu.RUnlock()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

// ComparePassword 校验明文密码与哈希是否匹配。
func (s *Service) ComparePassword(hash, password string) error {
	if hash == "" {
		return errors.New("hash is required")
	}
	if password == "" {
		return errors.New("password is required")
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Reload 更新密码服务配置。
func (s *Service) Reload(cfg Config) error {
	cost, err := normalizeCost(cfg.Cost)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cost = cost
	return nil
}

func normalizeCost(cost int) (int, error) {
	if cost == 0 {
		return bcrypt.DefaultCost, nil
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return 0, errors.New("invalid bcrypt cost")
	}
	return cost, nil
}
