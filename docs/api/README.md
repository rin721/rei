# API

The scaffold exposes these route groups:

- `GET /health`
- `GET /api/v1/samples`
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `POST /api/v1/auth/change-password`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `GET /api/v1/rbac/check`
- `POST /api/v1/rbac/roles/assign`
- `POST /api/v1/rbac/roles/revoke`
- `GET /api/v1/rbac/users/:user_id/roles`
- `GET /api/v1/rbac/roles/:role/users`
- `POST /api/v1/rbac/policies`
- `DELETE /api/v1/rbac/policies`
- `GET /api/v1/rbac/policies`

Response shape is always:

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "traceId": "trace-id",
  "serverTime": 1710000000
}
```
