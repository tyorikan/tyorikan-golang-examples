package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	// Logging the start and end of each request
	r.Use(middleware.Logger)

	// Path, Method 単位で定義
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	// Listen port
	http.ListenAndServe(":8080", r)
}
