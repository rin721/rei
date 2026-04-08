package rbac

import (
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const defaultModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = (g(r.sub, p.sub) || r.sub == p.sub) && r.obj == p.obj && r.act == p.act
`

// Config 描述 RBAC 配置。
type Config struct {
	ModelText string
	AutoSave  bool
}

// Policy 描述最小 RBAC 策略项。
type Policy struct {
	Subject string
	Object  string
	Action  string
}

// Manager 提供角色、策略与权限检查能力。
type Manager struct {
	mu       sync.RWMutex
	cfg      Config
	enforcer *casbin.Enforcer
}

// New 创建一个新的 RBAC 管理器。
func New(cfg Config) (*Manager, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	enforcer, err := newEnforcer(normalized)
	if err != nil {
		return nil, err
	}

	return &Manager{
		cfg:      normalized,
		enforcer: enforcer,
	}, nil
}

// CheckPermission 检查主体是否拥有指定权限。
func (m *Manager) CheckPermission(subject, object, action string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.enforcer.Enforce(subject, object, action)
}

// AssignRole 为用户分配角色。
func (m *Manager) AssignRole(user, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.enforcer.AddRoleForUser(user, role)
	return err
}

// RevokeRole 撤销用户角色。
func (m *Manager) RevokeRole(user, role string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.enforcer.DeleteRoleForUser(user, role)
	return err
}

// GetUserRoles 返回用户已分配角色。
func (m *Manager) GetUserRoles(user string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enforcer.GetRolesForUser(user)
}

// GetUsersForRole 返回指定角色下的用户列表。
func (m *Manager) GetUsersForRole(role string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enforcer.GetUsersForRole(role)
}

// AddPolicy 添加单条权限策略。
func (m *Manager) AddPolicy(subject, object, action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.enforcer.AddPolicy(subject, object, action)
	return err
}

// RemovePolicy 删除单条权限策略。
func (m *Manager) RemovePolicy(subject, object, action string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, err := m.enforcer.RemovePolicy(subject, object, action)
	return err
}

// AddPolicies 批量添加策略。
func (m *Manager) AddPolicies(policies ...Policy) error {
	if len(policies) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rules := make([][]string, 0, len(policies))
	for _, policy := range policies {
		rules = append(rules, []string{policy.Subject, policy.Object, policy.Action})
	}

	_, err := m.enforcer.AddPolicies(rules)
	return err
}

// RemovePolicies 批量删除策略。
func (m *Manager) RemovePolicies(policies ...Policy) error {
	if len(policies) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rules := make([][]string, 0, len(policies))
	for _, policy := range policies {
		rules = append(rules, []string{policy.Subject, policy.Object, policy.Action})
	}

	_, err := m.enforcer.RemovePolicies(rules)
	return err
}

// GetPolicies 返回当前所有权限策略。
func (m *Manager) GetPolicies() ([]Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rawPolicies, err := m.enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}
	policies := make([]Policy, 0, len(rawPolicies))
	for _, rule := range rawPolicies {
		if len(rule) < 3 {
			continue
		}
		policies = append(policies, Policy{
			Subject: rule[0],
			Object:  rule[1],
			Action:  rule[2],
		})
	}

	return policies, nil
}

// Reload 使用新配置重建 Enforcer，并保留已有角色和策略。
func (m *Manager) Reload(cfg Config) error {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	policies, err := m.enforcer.GetPolicy()
	if err != nil {
		return err
	}
	groupings, err := m.enforcer.GetGroupingPolicy()
	if err != nil {
		return err
	}

	next, err := newEnforcer(normalized)
	if err != nil {
		return err
	}

	if len(policies) > 0 {
		if _, err := next.AddPolicies(policies); err != nil {
			return err
		}
	}
	if len(groupings) > 0 {
		if _, err := next.AddGroupingPolicies(groupings); err != nil {
			return err
		}
	}

	m.cfg = normalized
	m.enforcer = next
	return nil
}

// Close 关闭 RBAC 管理器。
func (m *Manager) Close() error {
	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.ModelText == "" {
		cfg.ModelText = defaultModelText
	}
	return cfg, nil
}

func newEnforcer(cfg Config) (*casbin.Enforcer, error) {
	modelRef, err := model.NewModelFromString(cfg.ModelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(modelRef)
	if err != nil {
		return nil, err
	}
	enforcer.EnableAutoSave(cfg.AutoSave)
	return enforcer, nil
}
