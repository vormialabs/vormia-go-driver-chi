package chi

import (
	"context"
	"net/http"

	gochi "github.com/go-chi/chi/v5"
)

// Router wraps a go-chi mux and exposes a Vormia-friendly API.
// The concrete *Router will structurally satisfy the Router interface
// that vormia-go defines later — no framework import required.
type Router struct {
	mux    gochi.Router
	server *http.Server
}

// New creates a Router backed by go-chi.
func New() *Router {
	return &Router{mux: gochi.NewRouter()}
}

// --- HTTP verbs ---

func (r *Router) Get(pattern string, h http.HandlerFunc)     { r.mux.Get(pattern, h) }
func (r *Router) Post(pattern string, h http.HandlerFunc)    { r.mux.Post(pattern, h) }
func (r *Router) Put(pattern string, h http.HandlerFunc)     { r.mux.Put(pattern, h) }
func (r *Router) Patch(pattern string, h http.HandlerFunc)   { r.mux.Patch(pattern, h) }
func (r *Router) Delete(pattern string, h http.HandlerFunc)  { r.mux.Delete(pattern, h) }
func (r *Router) Head(pattern string, h http.HandlerFunc)    { r.mux.Head(pattern, h) }
func (r *Router) Options(pattern string, h http.HandlerFunc) { r.mux.Options(pattern, h) }

// Use registers global middleware. Must be called BEFORE routes are added.
func (r *Router) Use(mw ...Middleware) {
	for _, m := range mw {
		r.mux.Use(m)
	}
}

// Group runs fn on an inline sub-router that inherits middleware.
// No path prefix — use this to apply middleware to a set of routes.
func (r *Router) Group(fn func(sub *Router)) {
	r.mux.Group(func(c gochi.Router) {
		fn(&Router{mux: c})
	})
}

// Route mounts a sub-router under a path prefix (e.g. "/api").
func (r *Router) Route(prefix string, fn func(sub *Router)) {
	r.mux.Route(prefix, func(c gochi.Router) {
		fn(&Router{mux: c})
	})
}

// ServeHTTP makes *Router itself an http.Handler (handy for tests
// and for embedding behind other handlers).
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Handler returns the underlying handler.
func (r *Router) Handler() http.Handler {
	return r.mux
}

// Serve builds an *http.Server and blocks serving requests.
func (r *Router) Serve(addr string) error {
	r.server = &http.Server{
		Addr:    addr,
		Handler: r.mux,
	}
	return r.server.ListenAndServe()
}

// Shutdown gracefully stops the server started by Serve.
func (r *Router) Shutdown(ctx context.Context) error {
	if r.server == nil {
		return nil
	}
	return r.server.Shutdown(ctx)
}

// Routes returns every registered method+pattern pair, via chi.Walk.
// The return type matches vormia-go's contract.RouteInfo alias (anonymous
// struct) without importing the framework.
func (r *Router) Routes() []struct {
	Method  string
	Pattern string
} {
	var out []struct {
		Method  string
		Pattern string
	}
	_ = gochi.Walk(r.mux, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		out = append(out, struct {
			Method  string
			Pattern string
		}{Method: method, Pattern: route})
		return nil
	})
	return out
}

// URLParam extracts a path parameter (e.g. {id}) from the request.
// Re-exported so applications never import go-chi directly.
func URLParam(r *http.Request, key string) string {
	return gochi.URLParam(r, key)
}
