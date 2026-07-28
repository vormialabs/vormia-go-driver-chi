# Changelog

All notable changes to this project will be documented in this file.

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

[v1.0.0]: https://github.com/vormialabs/vormia-go-driver-chi/releases/tag/v1.0.0
