package middleware

import (
	"net/http"
	"strings"
)

// CORS lets the React frontend (a different origin) call this API from the
// browser. Only the origins in the allow-list receive CORS headers; "*" allows
// any origin (fine for demos, avoid it in production).
//
// It MUST wrap the router: browsers send a preflight OPTIONS request before
// POST/PUT/PATCH/DELETE with JSON, and the router has no OPTIONS routes.
func CORS(allowedOrigins []string) Middleware {
	allowAll := false
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		switch o {
		case "":
		case "*":
			allowAll = true
		default:
			allowed[o] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The response depends on the Origin header: tell caches about it.
			w.Header().Add("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin == "" { // not a browser cross-origin call (curl, Postman, health checks)
				next.ServeHTTP(w, r)
				return
			}

			_, ok := allowed[origin]
			preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""

			if !ok && !allowAll {
				if preflight {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r) // no CORS headers: the browser will block the response
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", origin)

			if preflight {
				h := w.Header()
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				h.Set("Access-Control-Max-Age", "86400") // browsers may cache the preflight for 24h
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Location")
			next.ServeHTTP(w, r)
		})
	}
}
