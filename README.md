# vormia-go-driver-chi

Chi router adapter for the Vormia Go framework. Thin wrapper around [go-chi/chi](https://github.com/go-chi/chi) that exposes an idiomatic HTTP routing API with middleware support — without leaking the underlying Chi types to your app.

**Version:** v1.1.0 — includes `Routes()` for vormia-go **v1.2.0+** `contract.Router`.

## Install

```bash
go get github.com/vormialabs/vormia-go-driver-chi@v1.1.0
```

## Quick start

```go
package main

import (
	"net/http"

	chi "github.com/vormialabs/vormia-go-driver-chi"
)

func main() {
	r := chi.New()
	r.Use(chi.CORS())

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = chi.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Route mounts a sub-router under a path prefix.
	r.Route("/api", func(api *chi.Router) {
		api.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
			_ = chi.Success(w, http.StatusOK, map[string]string{
				"id": chi.URLParam(req, "id"),
			})
		})
	})

	_ = r.Serve(":8080")
}
```

## API overview

### Router

| Method | Purpose |
|--------|---------|
| `New()` | Create a router (no args) |
| `Get` / `Post` / `Put` / `Patch` / `Delete` / `Head` / `Options` | Register routes |
| `Use(mw ...Middleware)` | Global middleware — call **before** routes |
| `Group(fn)` | Inline sub-router, inherits middleware, **no** path prefix |
| `Route(prefix, fn)` | Sub-router mounted under a path prefix (e.g. `"/api"`) |
| `Serve(addr)` | Start blocking `http.Server` |
| `Shutdown(ctx)` | Graceful shutdown |
| `ServeHTTP` / `Handler()` | Use as a plain `http.Handler` |
| `Routes()` | List registered method+pattern pairs (chi.Walk) |

### Helpers

- `JSON(w, status, data)` — raw JSON body
- `Success(w, status, data)` — `{"data": ...}`
- `Error(w, status, message)` — `{"error": "..."}`
- `URLParam(r, key)` — read `{id}`-style path params (re-exported so you never import go-chi)
- `CORS(origins...)` — permissive CORS middleware (default origin `*`)

### Path params

Chi uses `{id}` syntax, not `:id`:

```go
r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	_ = chi.JSON(w, http.StatusOK, map[string]string{"id": id})
})
```

### Group vs Route

```go
// Group: same path space, shared middleware only
r.Group(func(sub *chi.Router) {
	sub.Use(authMiddleware)
	sub.Get("/profile", profileHandler)
})

// Route: mounted under a prefix
r.Route("/api", func(api *chi.Router) {
	api.Get("/ping", pingHandler) // → GET /api/ping
})
```

## Example

See [`examples/basic`](examples/basic) for a runnable smoke-test server:

```bash
go run ./examples/basic
curl -i localhost:8080/health
curl -i localhost:8080/api/users/42
```

## Docs worth knowing

- [go-chi/chi](https://github.com/go-chi/chi) — routing, middleware order, `{param}` patterns
- [net/http](https://pkg.go.dev/net/http) — `Handler`, `HandlerFunc`, `Server`, graceful `Shutdown`
- [httptest](https://pkg.go.dev/net/http/httptest) — in-memory request testing

## License

MIT
