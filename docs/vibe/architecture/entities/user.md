# User Entity

更新日期: 2026-04-20

## 领域实体

文件: `internal/domain/user/entity.go`

```text
User
├─ ID
├─ Username
├─ Email
├─ DisplayName
├─ PasswordHash
├─ Status
├─ CreatedAt
└─ UpdatedAt
```

## 边界说明

- 这是纯领域实体，不带 `gorm` 标签。
- 这是纯领域实体，不带 `json` 标签。
- 它不依赖 Gin、GORM、JWT、RBAC 或任何基础设施库。
- 持久化形态由 `internal/models.User` 承担。
- HTTP 暴露形态由 `types/user` 中的 DTO 承担。
