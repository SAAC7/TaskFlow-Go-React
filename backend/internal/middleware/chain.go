// Package middleware contains the HTTP middlewares of the API:
// request ID, logging, panic recovery and CORS.
package middleware

import "net/http"

// Middleware wraps an http.Handler adding behaviour before/after it.
type Middleware func(http.Handler) http.Handler

// Chain applies the middlewares so that the FIRST one is the outermost:
// Chain(h, A, B, C) executes A -> B -> C -> h on the way in.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
