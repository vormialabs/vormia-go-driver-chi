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

	r.Route("/api", func(api *chi.Router) {
		api.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {
			_ = chi.Success(w, http.StatusOK, map[string]string{"id": chi.URLParam(req, "id")})
		})
	})

	_ = r.Serve(":8080")
}
