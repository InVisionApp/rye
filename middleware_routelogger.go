// This middleware provides simple logging output for the specific route
//
// You can use this middleware by specifying `rye.MiddlewareRouteLogger`
// when defining your routes.
//
// Example use case:
//
// ```
//  routes.Handle("/some/route", a.Dependencies.MWHandler.Handle([]rye.Handler{
//     rye.MiddlewareRouteLogger,
//     yourHandler,
// })).Methods("PUT", "OPTIONS")
// ```

package rye

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func MiddlewareRouteLogger() func(rw http.ResponseWriter, req *http.Request) *Response {
	return func(rw http.ResponseWriter, r *http.Request) *Response {
		log.Infof("%s \"%s %s %s\"", r.RemoteAddr, r.Method, r.RequestURI, r.Proto)
		return nil
	}
}
