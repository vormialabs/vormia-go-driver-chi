package chi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chi "github.com/vormialabs/vormia-go-driver-chi"
)

func do(t *testing.T, r *chi.Router, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func TestGet(t *testing.T) {
	r := chi.New()
	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_ = chi.JSON(w, http.StatusOK, map[string]string{"message": "Hello"})
	})

	rr := do(t, r, http.MethodGet, "/hello")
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rr.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), `"message":"Hello"`) {
		t.Fatalf("body: got %q", string(body))
	}
}

func TestURLParam(t *testing.T) {
	r := chi.New()
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
		_ = chi.JSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(req, "id")})
	})

	rr := do(t, r, http.MethodGet, "/users/123")
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), `"id":"123"`) {
		t.Fatalf("param not extracted: %q", string(body))
	}
}

func TestMiddlewareRuns(t *testing.T) {
	r := chi.New()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("X-Test", "on")
			next.ServeHTTP(w, req)
		})
	})
	r.Get("/x", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	rr := do(t, r, http.MethodGet, "/x")
	if rr.Header().Get("X-Test") != "on" {
		t.Fatal("middleware did not run")
	}
}

func TestRoutePrefix(t *testing.T) {
	r := chi.New()
	r.Route("/api", func(sub *chi.Router) {
		sub.Get("/ping", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	})

	if rr := do(t, r, http.MethodGet, "/api/ping"); rr.Code != http.StatusOK {
		t.Fatalf("prefixed route: got %d", rr.Code)
	}
	if rr := do(t, r, http.MethodGet, "/ping"); rr.Code != http.StatusNotFound {
		t.Fatalf("unprefixed should 404: got %d", rr.Code)
	}
}
