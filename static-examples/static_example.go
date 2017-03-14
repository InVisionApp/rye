package main

import (
	"fmt"
	"net/http"
	"os"

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

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("NewStaticFile: Could not get working directory.")
	}

	routes := mux.NewRouter().StrictSlash(true)

	routes.Handle("/", middlewareHandler.Handle([]rye.Handler{
		middlewareFirstHandler,
		homeHandler,
	})).Methods("GET")

	routes.PathPrefix("/dist/").Handler(middlewareHandler.Handle([]rye.Handler{
		rye.MiddlewareRouteLogger(),
		rye.NewStaticFilesystem(pwd+"/dist/", "/dist/"),
	}))

	routes.PathPrefix("/ui/").Handler(middlewareHandler.Handle([]rye.Handler{
		rye.MiddlewareRouteLogger(),
		rye.NewStaticFile(pwd + "/dist/index.html"),
	}))

	log.Infof("API server listening on %v", "localhost:8181")

	srv := &http.Server{
		Addr:    "localhost:8181",
		Handler: routes,
	}

	srv.ListenAndServe()
}

func middlewareFirstHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Middleware handler has fired!")
	return nil
}

func homeHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
	log.Infof("Home handler has fired!")

	fmt.Fprint(rw, "This is the home handler")
	return nil
}
