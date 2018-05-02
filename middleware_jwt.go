package rye

import "net/http"

const (
	CONTEXT_JWT = "rye-middlewarejwt-jwt"
)

type jwtVerify struct {
	secret string
	token  string
}

/*
This middleware is deprecated. Use NewMiddlewareAuth with NewJWTAuthFunc instead.

This remains here as a shim for backwards compatibility.

---------------------------------------------------------------------------

This middleware provides JWT verification functionality

You can use this middleware by specifying `rye.NewMiddlewareJWT(shared_secret)`
when defining your routes.

This middleware has no default version, it must be configured with a shared secret.

Example use case:

	routes.Handle("/some/route", a.Dependencies.MWHandler.Handle(
		[]rye.Handler{
			rye.NewMiddlewareJWT("this is a big secret"),
			yourHandler,
		})).Methods("PUT", "OPTIONS")

Additionally, this middleware puts the JWT token into the context for use by other
middlewares in your chain.

Access to that is simple (using the CONTEXT_JWT constant as a key)

	func getJWTfromContext(rw http.ResponseWriter, r *http.Request) *rye.Response {

		// Retrieving the value is easy!
		// Just reference the rye.CONTEXT_JWT const as a key
		myVal := r.Context().Value(rye.CONTEXT_JWT)

		// Log it to the server log?
		log.Infof("Context Value: %v", myVal)

		return nil
	}

*/
func NewMiddlewareJWT(secret string) func(rw http.ResponseWriter, req *http.Request) *Response {
	return NewMiddlewareAuth(NewJWTAuthFunc(secret))
}
