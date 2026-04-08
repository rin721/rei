package config

// RBACConfig 描述 RBAC 配置。
type RBACConfig struct {
	Enabled     bool   `yaml:"enabled" env:"RBAC_ENABLED"`
	ModelPath   string `yaml:"model_path" env:"RBAC_MODEL_PATH"`
	PolicyTable string `yaml:"policy_table" env:"RBAC_POLICY_TABLE"`
	AutoSave    bool   `yaml:"auto_save" env:"RBAC_AUTO_SAVE"`
}

// ValidateName 返回配置域名。
func (c RBACConfig) ValidateName() string {
	return "rbac"
}

// ValidateRequired 返回该配置域是否必需。
func (c RBACConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 RBACConfig。
func (c RBACConfig) Validate() error {
	return nil
}
