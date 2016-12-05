/*
Package Rye is a simple library to support http services.
Rye provides a middleware handler which can be used to chain http handlers together while providing
simple statsd metrics for use with a monitoring solution such as DataDog or other logging aggregators.
Rye also provides some additional middleware handlers that are entirely optional but easily consumed using Rye.

Setup

In order to use rye, you should vendor it and the statsd client within your project.

	govendor fetch github.com/cactus/go-statsd-client/statsd

	# Rye is a private repo, so we should clone it first
	mkdir -p $GOPATH/github.com/InVisionApp
	cd $GOPATH/github.com/InVisionApp
	git clone git@github.com:InVisionApp/rye.git

	govendor add github.com/InVisionApp/rye

Writing custom middleware handlers

Begin by importing the required libraries:
	import (
		"github.com/cactus/go-statsd-client/statsd"
		"github.com/InVisionApp/rye"
	)
Create a statsd client (if desired) and create a rye Config in order to pass in optional dependencies:
	config := &rye.Config{
		Statter:	statsdClient,
		StatRate:	DEFAULT_STATSD_RATE,
	}
Create a middleware handler. The purpose of the Handler is to keep Config and to provide an interface for chaining http handlers.
	middlewareHandler := rye.NewMWHandler(config)
Build your http handlers using the Handler type from the **rye** package.
	type Handler func(w http.ResponseWriter, r *http.Request) *rye.Response
Here are some example (custom) handlers:
	func homeHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
		fmt.Fprint(rw, "Refer to README.md for auth-api API usage")
		return nil
	}

	func middlewareFirstHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
		fmt.Fprint(rw, "This handler fires first.")
		return nil
	}

	func errorHandler(rw http.ResponseWriter, r *http.Request) *rye.Response {
		return &rye.Response {
			StatusCode: http.StatusInternalServerError,
			Err:        errors.New(message),
		}
	}
Finally, to setup your handlers in your API
	routes := mux.NewRouter().StrictSlash(true)

	routes.Handle("/", middlewareHandler.Handle(
		[]rye.Handler{
			a.middlewareFirstHandler,
			a.homeHandler,
		})).Methods("GET")

	log.Infof("API server listening on %v", ListenAddress)

	srv := &http.Server{
		Addr:		ListenAddress,
		Handler:	routes,
	}

	srv.ListenAndServe()

Statsd Generated by Rye

Rye comes with built-in configurable `statsd` statistics that you could record to your favorite monitoring system. To configure that, you'll need to set up a `Statter` based on the `github.com/cactus/go-statsd-client` and set it in your instantiation of `MWHandler` through the `rye.Config`.

When a middleware is called, it's timing is recorded and a counter is recorded associated directly with the http status code returned during the call. Additionally, an `errors` counter is also sent to the statter which allows you to count any errors that occur with a code equaling or above 500.

Example: If you have a middleware handler you've created with a method named `loginHandler`, successful calls to that will be recorded to `handlers.loginHandler.2xx`. Additionally you'll receive stats such as `handlers.loginHandler.400` or `handlers.loginHandler.500`. You also will receive an increase in the `errors` count.

If you're sending your logs into a system such as DataDog, be aware that your stats from Rye can have prefixes such as `statsd.my-service.my-k8s-cluster.handlers.loginHandler.2xx` or even `statsd.my-service.my-k8s-cluster.errors`. Just keep in mind your stats could end up in the destination sink system with prefixes.

Using built-in middleware handlers

Rye comes with various pre-built middleware handlers. Pre-built middlewares  source (and docs) can be found in the package dir following the pattern `middleware_*.go`.

To use them, specify the constructor of the middleware as one of the middleware handlers when you define your routes:
	// example
	routes.Handle("/", middlewareHandler.Handle(
		[]rye.Handler{
			rye.MiddlewareCORS(), // to use the CORS middleware (with defaults)
			a.homeHandler,
		})).Methods("GET")
OR
	routes.Handle("/", middlewareHandler.Handle(
		[]rye.Handler{
			rye.NewMiddlewareCORS("*", "GET, POST", "X-Access-Token"), // to use specific config when instantiating the middleware handler
			a.homeHandler,
		})).Methods("GET")

 */
package rye


