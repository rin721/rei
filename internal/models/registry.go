package models

// All 返回所有需要纳入迁移管理的 Model 实例（指针）。
//
// db generate 命令通过此函数获取 Model 列表并反射生成 DDL。
// 每次新增或删除业务 Model 时，同步更新此列表。
func All() []any {
	return []any{
		&User{},
		&Role{},
		&UserRole{},
		&Policy{},
		&Sample{},
	}
}
