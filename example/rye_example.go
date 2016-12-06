package main

import (
	"errors"
	"fmt"
	"net/http"
	"context"

	"github.com/InVisionApp/rye"
	log "github.com/Sirupsen/logrus"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gorilla/mux"
)

func main() {
	statsdClient, err := statsd.NewBufferedClient("localhost:12345", "my_service", 1.0, 0)
	if err != nil {
		log.Fatalf("Unable to instantiate statsd client: %v", err.Error())
	}

	config := rye.Config{
		Statter:  statsdClient,
		StatRate: 1.0,
	}

	middlewareHandler := rye.NewMWHandler(config)

	routes := mux.NewRouter().StrictSlash(true)

	routes.Handle("/", middlewareHandler.Handle([]rye.Handler{
		middlewareFirstHandler,
		homeHandler,
	})).Methods("GET")

	// If you perform a `curl -i http://localhost:8181/cors -H "Origin: *.foo.com"`
	// you will see that the CORS middleware is adding required headers
	routes.Handle("/cors", middlewareHandler.Handle([]rye.Handler{
		rye.MiddlewareCORS(),
		homeHandler,
	})).Methods("GET", "OPTIONS")

	// If you perform an `curl -i http://localhost:8181/jwt \
	// -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
	// you will see that we are allowed through to the handler, if the sample token is changed, we will get a 401
	routes.Handle("/jwt", middlewareHandler.Handle([]rye.Handler{
		rye.NewMiddlewareJWT("secret"),
		homeHandler,
	})).Methods("GET")

	routes.Handle("/error", middlewareHandler.Handle([]rye.Handler{
		middlewareFirstHandler,
		errorHandler,
		homeHandler,
	})).Methods("GET")

	routes.Handle("/context", middlewareHandler.Handle(
		[]rye.Handler{
			stashContextHandler,
			logContextHandler,
	})).Methods("GET")

	log.Infof("API server listening on %v", "localhost:8181")

	srv := &http.Server{
		Addr:    "localhost:8181",
		Handler: routes,
	}

	srv.ListenAndServe()
}

func homeHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Home handler has fired!")

	fmt.Fprint(rw, "This is the home handler")
	return nil
}

func middlewareFirstHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Middleware handler has fired!")
	return nil
}

func errorHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Error handler has fired!")

	message := "This is the error handler"

	return &rye.Response{
		StatusCode: http.StatusInternalServerError,
		Err:        errors.New(message),
	}
}

func stashContextHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Stash Context handler has fired!")
	ctx := r.Context()

	toContext := r.URL.Query().Get("ctx")
	if toContext != "" {
		log.Infof("Adding `query-string-ctx` to request.Context(). Val: %v",toContext)
	} else {
		log.Infof("Adding default `query-string-ctx` value to context")
		toContext = "No value added. Add querystring param `ctx` with a value to get it mirrored through context."
	}
	ctx = context.WithValue(ctx,"query-string-ctx",toContext)
	log.Infof("Me")
	return &rye.Response{Context:ctx}
}

func logContextHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Log Context handler has fired!")
	fromContext := r.Context().Value("query-string-ctx")
	log.Infof("From context: %v",fromContext)
	fmt.Fprintf(rw,"Here's the `ctx` query string value you passed. Pulled from context: %v",fromContext)
	return nil
}