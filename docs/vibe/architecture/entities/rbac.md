# RBAC Entities

Updated On: 2026-04-20

## Domain Entities

Files:
- `internal/domain/rbac/entities.go`

```text
Role
├─ ID
├─ Name
└─ Description

RoleBinding
├─ ID
├─ UserID
└─ RoleName

Policy
├─ ID
├─ Subject
├─ Object
└─ Action
```

## Boundary Notes

- These are pure domain entities with no `gorm` tags.
- These are pure domain entities with no `json` tags.
- They do not depend on Gin, GORM, JWT, or the runtime RBAC manager.
- Persistence shape remains in `internal/models`.
- HTTP shape remains in `types`.
