# transport/http

HTTP transport layer (Gin handlers + middleware).

- Handlers translate HTTP ↔ application DTOs only.
- Register protected admin routes with `middleware.AdminAuth`.
- Copy `example` for a new resource surface.
