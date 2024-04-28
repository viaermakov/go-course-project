package transport

import (
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Method: %s Path: %s", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
