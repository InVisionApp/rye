// This middleware provides CORS functionality
//
// Example use case:
//
//  routes.Handle("/some/route", a.Dependencies.MWHandler.Handle([]rye.Handler{
//     rye.middlewares.NewCORS(),
//     yourHandler,
// })).Methods("PUT", "OPTIONS")

package middlewares

import (
	"net/http"
)

const (
	DEFAULT_CORS_ALLOW_ORIGIN  = "*"
	DEFAULT_CORS_ALLOW_METHODS = "POST, GET, OPTIONS, PUT, DELETE"
	DEFAULT_CORS_ALLOW_HEADERS = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Access-Token"
)

type CORS struct {
	CORSAllowOrigin  string
	CORSAllowMethods string
	CORSAllowHeaders string
}

func NewCORSMiddleware() func(rw http.ResponseWriter, req *http.Request) *Response {
	c := &CORS{
		CORSAllowOrigin:  DEFAULT_CORS_ALLOW_ORIGIN,
		CORSAllowMethods: DEFAULT_CORS_ALLOW_METHODS,
		CORSAllowHeaders: DEFAULT_CORS_ALLOW_HEADERS,
	}

	return c.handle
}

func NewCORSMiddlewareParams(origin, methods, headers string) func(rw http.ResponseWriter, req *http.Request) *Response {
	c := &CORS{
		CORSAllowOrigin:  origin,
		CORSAllowMethods: methods,
		CORSAllowHeaders: headers,
	}

	return c.handle
}

// If `Origin` header gets passed, add required response headers for CORS support.
// Return bool if `Origin` header was detected.
func (c *CORS) handle(rw http.ResponseWriter, req *http.Request) *Response {
	origin := req.Header.Get("Origin")

	// Origin header not provided, nothing for CORS to do
	if origin == "" {
		return nil
	}

	rw.Header().Set("Access-Control-Allow-Origin", c.CORSAllowOrigin)
	rw.Header().Set("Access-Control-Allow-Methods", c.CORSAllowMethods)
	rw.Header().Set("Access-Control-Allow-Headers", c.CORSAllowHeaders)

	// If this was a preflight request, stop further middleware execution
	if req.Method == "OPTIONS" {
		return &Response{
			StopExecution: true,
		}
	}

	return nil
}
