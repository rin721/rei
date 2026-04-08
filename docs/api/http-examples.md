# HTTP Examples

Register a user:

```bash
curl -X POST http://127.0.0.1:9999/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"Password123","displayName":"Alice"}'
```

Login:

```bash
curl -X POST http://127.0.0.1:9999/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"Password123"}'
```

Fetch the current user profile:

```bash
curl http://127.0.0.1:9999/api/v1/users/me \
  -H "Authorization: Bearer <access-token>"
```

Check RBAC permission:

```bash
curl "http://127.0.0.1:9999/api/v1/rbac/check?object=/api/v1/users/me&action=GET" \
  -H "Authorization: Bearer <access-token>"
```
