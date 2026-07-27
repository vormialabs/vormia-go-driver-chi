package chi

import "net/http"

// Middleware is the standard net/http middleware signature.
// It's an alias (not a new type) so it's interchangeable with plain
// func(http.Handler) http.Handler everywhere.
type Middleware = func(http.Handler) http.Handler

// CORS returns a permissive CORS middleware. Pass a single origin to
// restrict it; default is "*".
func CORS(origins ...string) Middleware {
	origin := "*"
	if len(origins) > 0 {
		origin = origins[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
