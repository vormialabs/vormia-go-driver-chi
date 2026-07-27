# How `vormia-go-driver-chi` Works

A walkthrough of what this package does, how each file fits together, and the design decisions behind it.

## What this package is

`vormia-go-driver-chi` is a **thin adapter** around [go-chi/chi v5](https://github.com/go-chi/chi). It gives Vormia applications a small, stable routing API (`Get`, `Post`, `Use`, `Route`, `Serve`, …) **without ever exposing Chi's types** to the application. That means:

- Your app imports only this package (`chi "github.com/vormialabs/vormia-go-driver-chi"`) — never `go-chi/chi` directly.
- The Vormia core framework can later define a `Router` interface, and this package's `*Router` will satisfy it **structurally** (Go interfaces are implicit — no `implements` keyword needed, so this package needs zero imports from vormia-go).
- If Vormia ever swaps Chi for another mux, application code doesn't change — only the driver does.

The whole package is ~150 lines across three files:

| File | Responsibility |
|------|----------------|
| `router.go` | The `Router` struct: verbs, grouping, serving, shutdown, `URLParam` |
| `middleware.go` | The `Middleware` type alias and the built-in `CORS()` middleware |
| `response.go` | JSON response helpers: `JSON`, `Success`, `Error` |

## The core: `router.go`

### The wrapper struct

```13:16:router.go
type Router struct {
	mux    gochi.Router
	server *http.Server
}
```

`Router` holds two things:

- `mux` — the actual Chi router that does all the pattern matching. It's an unexported field, which is the whole trick: nothing outside this package can touch Chi directly.
- `server` — the `*http.Server` created by `Serve()`, kept around so `Shutdown()` can stop it gracefully later.

`New()` just creates a Chi mux and wraps it. No config arguments — configuration happens through methods (`Use`, etc.).

### HTTP verbs are one-line delegations

```25:31:router.go
func (r *Router) Get(pattern string, h http.HandlerFunc)     { r.mux.Get(pattern, h) }
func (r *Router) Post(pattern string, h http.HandlerFunc)    { r.mux.Post(pattern, h) }
func (r *Router) Put(pattern string, h http.HandlerFunc)     { r.mux.Put(pattern, h) }
func (r *Router) Patch(pattern string, h http.HandlerFunc)   { r.mux.Patch(pattern, h) }
func (r *Router) Delete(pattern string, h http.HandlerFunc)  { r.mux.Delete(pattern, h) }
func (r *Router) Head(pattern string, h http.HandlerFunc)    { r.mux.Head(pattern, h) }
func (r *Router) Options(pattern string, h http.HandlerFunc) { r.mux.Options(pattern, h) }
```

Handlers are plain `http.HandlerFunc` from the standard library — no custom handler type, no context wrapper. Everything you already know about `net/http` applies directly.

### Middleware: `Use()`

```34:38:router.go
func (r *Router) Use(mw ...Middleware) {
	for _, m := range mw {
		r.mux.Use(m)
	}
}
```

**Important Chi rule:** `Use` must be called **before** any routes are registered on that router. Chi panics if you add middleware after a route exists. That's why the README and example always do `r.Use(...)` first.

### `Group` vs `Route` — the two ways to nest

Both create a **sub-router** and hand it to your callback, but they differ in one thing: the path prefix.

```go
// Group: NO path prefix. Use it to apply middleware to a subset of routes.
r.Group(func(sub *chi.Router) {
	sub.Use(authMiddleware)       // only these routes get auth
	sub.Get("/profile", handler)  // still served at /profile
})

// Route: mounts under a prefix.
r.Route("/api", func(api *chi.Router) {
	api.Get("/ping", handler)     // served at /api/ping
})
```

Internally both wrap Chi's own `Group`/`Route` and re-wrap the child Chi router in a fresh `*Router`:

```42:46:router.go
func (r *Router) Group(fn func(sub *Router)) {
	r.mux.Group(func(c gochi.Router) {
		fn(&Router{mux: c})
	})
}
```

Note the sub-router only gets a `mux` — no `server`. Only the top-level router ever owns the HTTP server.

### Serving and graceful shutdown

- `Serve(addr)` builds an `*http.Server`, stores it on the struct, and calls `ListenAndServe()` — this **blocks** until the server stops.
- `Shutdown(ctx)` calls the stored server's `Shutdown`, which stops accepting new connections and waits for in-flight requests to finish (bounded by the context deadline). It's a no-op (returns `nil`) if `Serve` was never called.
- `ServeHTTP(w, req)` makes `*Router` itself an `http.Handler`. This is what makes in-memory testing possible (see below) and lets you embed the router behind other handlers or your own `http.Server`.
- `Handler()` returns the raw underlying handler if you want to wire it up yourself.

### `URLParam` — the one re-export

```85:87:router.go
func URLParam(r *http.Request, key string) string {
	return gochi.URLParam(r, key)
}
```

Chi stores matched path parameters in the request context. Chi's syntax is `{id}` (curly braces), **not** `:id` like Express or Laravel routes:

```go
r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id") // "42" for GET /users/42
})
```

This function is re-exported here purely so applications never need a `go-chi` import — keeping the abstraction airtight.

## Middleware: `middleware.go`

### The `Middleware` type

```8:8:middleware.go
type Middleware = func(http.Handler) http.Handler
```

Note the `=`: this is a **type alias**, not a new named type. That's deliberate — any standard `func(http.Handler) http.Handler` middleware (from Chi's middleware package, or anywhere in the Go ecosystem) can be passed to `Use()` without conversion, and vice versa.

The pattern itself is the standard Go "onion": a middleware receives the `next` handler and returns a new handler that does work before/after calling `next.ServeHTTP(w, r)`.

### Built-in `CORS()`

`CORS(origins...)` returns a permissive CORS middleware:

- No arguments → `Access-Control-Allow-Origin: *`
- `CORS("https://app.example.com")` → restricted to that single origin (only the first argument is used)
- It sets allowed methods and `Content-Type`/`Authorization` headers, and short-circuits `OPTIONS` preflight requests with `204 No Content` so they never reach your handlers.

## Responses: `response.go`

Three small helpers, all built on the first:

| Helper | Body shape | Use for |
|--------|-----------|---------|
| `JSON(w, status, data)` | exactly what you pass | raw payloads |
| `Success(w, status, data)` | `{"data": ...}` | consistent success envelope |
| `Error(w, status, message)` | `{"error": "..."}` | consistent error envelope |

`JSON` sets the `Content-Type` header, writes the status code, then streams the encoded body:

```12:16:response.go
func JSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
```

**Gotcha:** `WriteHeader(status)` is called *before* encoding. If `json.Encode` fails midway (e.g. an unmarshalable value like a channel), the status is already sent and can't be changed. If that matters for a handler, marshal to bytes first and only write on success.

## Request lifecycle, end to end

For `GET /api/users/42` in the example app:

1. `http.Server` (created by `Serve(":8080")`) accepts the connection and calls the Chi mux's `ServeHTTP`.
2. Global middleware runs first — here `CORS()` sets the headers, and since the method isn't `OPTIONS`, calls `next`.
3. Chi matches `/api` → descends into the sub-router mounted by `Route("/api", ...)`.
4. Chi matches `/users/{id}`, stores `id=42` in the request context.
5. Your handler runs, reads the param with `chi.URLParam(req, "id")`, and writes the response with `chi.Success(...)` → `{"data":{"id":"42"}}`.

## Testing without a network

Because `*Router` implements `http.Handler`, tests use `httptest` to drive requests entirely in memory — no port, no goroutines, no cleanup:

```13:19:router_test.go
func do(t *testing.T, r *chi.Router, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}
```

`httptest.NewRequest` fabricates a request, `httptest.NewRecorder` is a fake `ResponseWriter` that captures status/headers/body, and calling `r.ServeHTTP` runs the full middleware + routing pipeline synchronously. Run the suite with:

```bash
go test ./...
```

## Things to remember

- Call `Use()` **before** registering routes — Chi panics otherwise.
- Path params use `{id}` syntax, not `:id`.
- `Group` = shared middleware, same paths. `Route` = path prefix.
- `Serve()` blocks; call it last in `main` (or in a goroutine if you need to coordinate `Shutdown`).
- Handlers and middleware are 100% standard `net/http` — anything from the Go ecosystem plugs in.

## Further reading (official docs)

- [go-chi/chi](https://github.com/go-chi/chi) — routing patterns, middleware ordering, sub-routers
- [net/http](https://pkg.go.dev/net/http) — `Handler`, `HandlerFunc`, `Server.Shutdown`
- [net/http/httptest](https://pkg.go.dev/net/http/httptest) — in-memory request testing
- [encoding/json](https://pkg.go.dev/encoding/json) — `Encoder` behavior used by the response helpers
- [Effective Go: interfaces](https://go.dev/doc/effective_go#interfaces) — why structural typing lets this driver satisfy a future vormia-go interface with no import
