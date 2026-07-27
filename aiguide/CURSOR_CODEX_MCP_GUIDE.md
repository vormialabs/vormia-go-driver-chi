# AI Editor Guide — `vormia-go-driver-chi`

> Context file for AI coding tools (Cursor, Codex, Copilot, MCP-based agents).
> Read this first to understand and work with the package without exploring the whole repo.
> Human-oriented deep dive: [`aiguide/GUIDE.md`](GUIDE.md). Usage docs: [`README.md`](../README.md).

## Package facts

| Fact | Value |
|------|-------|
| Module | `github.com/vormialabs/vormia-go-driver-chi` |
| Go package name | `chi` (import as `chi "github.com/vormialabs/vormia-go-driver-chi"`) |
| Go version | 1.26.4 |
| Only dependency | `github.com/go-chi/chi/v5` |
| Purpose | Thin adapter around go-chi so Vormia apps never import go-chi directly |
| Size | ~150 lines, 3 source files |
| Test command | `go test ./...` |
| Runnable example | `go run ./examples/basic` |

## File map

| File | Contains |
|------|----------|
| `router.go` | `Router` struct, `New()`, HTTP verbs, `Use`, `Group`, `Route`, `Serve`, `Shutdown`, `ServeHTTP`, `Handler`, `URLParam` |
| `middleware.go` | `Middleware` type alias, `CORS()` |
| `response.go` | `JSON`, `Success`, `Error` response helpers |
| `router_test.go` | In-memory `httptest` suite |
| `examples/basic/` | Runnable smoke-test server |

## Full public API (exact signatures)

```go
// router.go
func New() *Router

func (r *Router) Get(pattern string, h http.HandlerFunc)
func (r *Router) Post(pattern string, h http.HandlerFunc)
func (r *Router) Put(pattern string, h http.HandlerFunc)
func (r *Router) Patch(pattern string, h http.HandlerFunc)
func (r *Router) Delete(pattern string, h http.HandlerFunc)
func (r *Router) Head(pattern string, h http.HandlerFunc)
func (r *Router) Options(pattern string, h http.HandlerFunc)

func (r *Router) Use(mw ...Middleware)              // call BEFORE registering routes
func (r *Router) Group(fn func(sub *Router))        // sub-router, NO path prefix
func (r *Router) Route(prefix string, fn func(sub *Router)) // sub-router under prefix

func (r *Router) Serve(addr string) error           // blocking ListenAndServe
func (r *Router) Shutdown(ctx context.Context) error // graceful; nil if Serve never ran
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) // *Router is an http.Handler
func (r *Router) Handler() http.Handler             // raw underlying handler

func URLParam(r *http.Request, key string) string   // read {id}-style path params

// middleware.go
type Middleware = func(http.Handler) http.Handler   // ALIAS, interchangeable with plain func
func CORS(origins ...string) Middleware             // no args => "*", else origins[0] only

// response.go
func JSON(w http.ResponseWriter, status int, data any) error      // raw body
func Success(w http.ResponseWriter, status int, data any) error   // {"data": ...}
func Error(w http.ResponseWriter, status int, message string) error // {"error": "..."}
```

There is no other public API. `Router.mux` and `Router.server` are unexported by design — never expose go-chi types to callers.

## Hard rules (violating these breaks code)

1. **`Use()` must be called before any route is registered** on the same router. go-chi panics otherwise.
2. **Path params use `{id}` syntax**, never `:id` (Express/Laravel style). Read them with `chi.URLParam(req, "id")`.
3. **Never import `github.com/go-chi/chi/v5` in application code.** The whole point of this driver is that only this package touches go-chi. If a feature needs go-chi, add a wrapper/re-export here instead.
4. **Handlers are plain `http.HandlerFunc`** — no custom handler type, no context wrapper. Anything from the `net/http` ecosystem plugs in directly.
5. **`Serve()` blocks.** Call it last in `main`, or in a goroutine when you need to coordinate `Shutdown(ctx)`.
6. **`Group` vs `Route`:** `Group` = shared middleware, same path space. `Route` = mounts under a path prefix. Sub-routers created by both have no `server` field; only the top-level router owns the `http.Server`.
7. **`JSON` writes the status before encoding.** If encoding can fail (unusual value types), marshal to bytes first and only then write.

## Canonical usage

```go
package main

import (
	"net/http"

	chi "github.com/vormialabs/vormia-go-driver-chi"
)

func main() {
	r := chi.New()
	r.Use(chi.CORS()) // middleware first, always

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		_ = chi.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

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

## Recipes

### Protected route group (middleware without prefix)

```go
r.Group(func(sub *chi.Router) {
	sub.Use(authMiddleware)          // applies only inside this group
	sub.Get("/profile", profileHandler) // still served at /profile
})
```

### Custom middleware

```go
// Any func(http.Handler) http.Handler works — Middleware is an alias.
func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before
		next.ServeHTTP(w, r)
		// after
	})
}
r.Use(logger)
```

### Graceful shutdown

```go
go func() { _ = r.Serve(":8080") }()
// ... on signal:
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
_ = r.Shutdown(ctx)
```

### Testing (no network, no ports)

```go
req := httptest.NewRequest(http.MethodGet, "/health", nil)
rr := httptest.NewRecorder()
r.ServeHTTP(rr, req) // runs full middleware + routing pipeline in memory
// assert on rr.Code, rr.Body, rr.Header()
```

### Embedding behind your own server

```go
srv := &http.Server{Addr: ":8080", Handler: r} // *Router is an http.Handler
// or: mux.Handle("/", r.Handler())
```

## Response envelope conventions

| Helper | Body written | When to use |
|--------|--------------|-------------|
| `JSON(w, 200, payload)` | `payload` as-is | raw/custom shapes |
| `Success(w, 200, payload)` | `{"data": payload}` | success responses in Vormia apps |
| `Error(w, 404, "not found")` | `{"error": "not found"}` | error responses in Vormia apps |

All three set `Content-Type: application/json` and return the `json.Encoder` error (typically ignored with `_ =`).

## When modifying this package

- Keep it thin: every method should be a one-or-few-line delegation to go-chi or `net/http`. No business logic.
- Preserve structural compatibility: vormia-go will later define a `Router` interface that `*Router` must satisfy implicitly — do not change existing method signatures.
- New go-chi features must be wrapped or re-exported (like `URLParam`), never surfaced as go-chi types.
- Sub-routers in `Group`/`Route` are wrapped as `&Router{mux: c}` — follow that pattern for any new nesting method.
- Add tests in `router_test.go` using the in-memory `do(t, r, method, path)` helper; run `go test ./...` before finishing.
