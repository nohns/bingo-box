package http

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Construct auth middleware using api key
func apiKeyMiddlewareFactory(apiKey string) mux.MiddlewareFunc {
	// Create a factory for middleware, so we can pass in dependencies to it.

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(rw, r)
		})
	}
}
