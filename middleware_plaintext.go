package tracker

import "net/http"

// Middleware for setting Content-Type to charset=ISO-8859-1.
func PlaintextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=ISO-8859-1")
		next.ServeHTTP(w, r)
	})
}
