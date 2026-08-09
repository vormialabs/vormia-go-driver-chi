# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0] - 2026-08-09

Adds **`Routes()`** so the router satisfies vormia-go **v1.2.0** `contract.Router` (route introspection for a future `route:list` CLI).

### Added

- **`Routes()`** — returns every registered method+pattern pair via [chi.Walk](https://pkg.go.dev/github.com/go-chi/chi/v5#Walk). Return type matches vormia-go's `RouteInfo` alias (anonymous struct) without importing the framework.

### Notes

- Pair with [vormia-go](https://github.com/vormialabs/vormia-go) **v1.2.0+**.
- Still requires `{id}` path parameter syntax (Chi convention), not `:id`.

## [v1.0.0] - 2026-07-27

Initial stable release of the Chi router adapter for the Vormia Go framework — a thin wrapper around [go-chi/chi](https://github.com/go-chi/chi) that exposes an idiomatic HTTP routing API without leaking the underlying Chi types to your app.

### Added

- **Router** (`New()`) with route registration for all common HTTP verbs: `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`.
- **Middleware support** via `Use(mw ...Middleware)` for global middleware.
- **Sub-routers**:
  - `Group(fn)` — inline sub-router that inherits middleware, with no path prefix.
  - `Route(prefix, fn)` — sub-router mounted under a path prefix (e.g. `"/api"`).
- **Server lifecycle**: blocking `Serve(addr)` and graceful `Shutdown(ctx)`.
- **`http.Handler` interoperability** via `ServeHTTP` and `Handler()`.
- **Response helpers**:
  - `JSON(w, status, data)` — raw JSON body.
  - `Success(w, status, data)` — wraps payload as `{"data": ...}`.
  - `Error(w, status, message)` — wraps message as `{"error": "..."}`.
- **Path params**: `URLParam(r, key)` re-exported so apps never import go-chi directly (`{id}`-style patterns).
- **CORS middleware**: `CORS(origins...)` with permissive default (`*`).
- Runnable example server in [`examples/basic`](examples/basic).
- Test suite (`router_test.go`) covering the routing API.

### Notes

- Requires the `{id}` path parameter syntax (Chi convention), not `:id`.
- Licensed under MIT.

[v1.1.0]: https://github.com/vormialabs/vormia-go-driver-chi/releases/tag/v1.1.0
[v1.0.0]: https://github.com/vormialabs/vormia-go-driver-chi/releases/tag/v1.0.0
