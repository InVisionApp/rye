
<img align="right" src="rye.gif">

# rye
A simple library to support http services. Currently, **rye** provides a middleware handler which can be used to chain http handlers together while providing statsd metrics for use with DataDog or other logging aggregators and out-of-the-box CORS support.

## Setup
In order to use **rye**, you should vendor it and the **statsd** client within your project.

```
govendor fetch github.com/cactus/go-statsd-client/statsd

# Rye is a private repo, so we should clone it first
mkdir -p $GOPATH/github.com/InVisionApp
cd $GOPATH/github.com/InVisionApp
git clone git@github.com:InVisionApp/rye.git

govendor add github.com/InVisionApp/rye
```

## Why another middleware lib?

* For one, `rye` is *tiny* - the entire library is <120 lines of code including comments
* In addition, each middleware gets statsd metrics tracking for free
* We also wanted to have an easy way to say “run these 2 middlewares on this endpoint, but only one middleware on this endpoint” 
    * Of course, this is doable with negroni and gorilla-mux, but you’d have to use a subrouter with gorilla, which tends to end up in more code
* Also, as a bonus, we bundled in some helper methods for standardizing JSON response messages
* And finally, we created a unified way for handlers and middlewares to return detailed errors (if they chose to do so)
* Oh yeah and it has built-in support for CORS too!

## Example
Begin by importing the required libraries:

```go
import (
    "github.com/cactus/go-statsd-client/statsd"
    "github.com/InVisionApp/rye"
)
```

Create a statsd client (if desired) and create a rye Config in order to pass in optional dependencies:

```go
config := &rye.Config{
        Statter:          statsdClient,
        StatRate:         DEFAULT_STATSD_RATE,
}
```

Create a middleware handler. The purpose of the Handler is to keep Config and to provide an interface for chaining http handlers.
```go
middlewareHandler := rye.NewMWHandler(config)
```

Build your http handlers using the Handler type from **rye**.

```go
type Handler func(w http.ResponseWriter, r *http.Request) *DetailedError
```

Here are some example handlers:

```go
func homeHandler(rw http.ResponseWriter, r *http.Request) *rye.DetailedError {
	fmt.Fprint(rw, "Refer to README.md for auth-api API usage")
	return nil
}

func middlewareFirstHandler(rw http.ResponseWriter, r *http.Request) *rye.DetailedError {
	fmt.Fprint(rw, "This handler fires first.")
	return nil
}

func errorHandler(rw http.ResponseWriter, r *http.Request) *rye.DetailedError {
	return &rye.DetailedError{
    			StatusCode: http.StatusInternalServerError,
    			Err:        errors.New(message),
    }
}
```

Finally, to setup your handlers in your api (Example shown using [Gorilla](https://github.com/gorilla/mux)):
```go
routes := mux.NewRouter().StrictSlash(true)

routes.Handle("/", middlewareHandler.Handle([]rye.Handler{
    a.middlewareFirstHandler,
    a.homeHandler,
})).Methods("GET")

log.Infof("API server listening on %v", ListenAddress)

srv := &http.Server{
    Addr:         ListenAddress,
    Handler:      routes,
    ReadTimeout:  time.Duration(ReadTimeout) * time.Second,
    WriteTimeout: time.Duration(WriteTimeout) * time.Second,
}

srv.ListenAndServe()

```
## Full Example
```go
package main

import (
    "errors"
    "fmt"
    "net/http"

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

    config := rye.Config{statsdClient, 1.0}

    middlewareHandler := rye.NewMWHandler(config)

    routes := mux.NewRouter().StrictSlash(true)

    routes.Handle("/", middlewareHandler.Handle([]rye.Handler{
        middlewareFirstHandler,
        homeHandler,
    })).Methods("GET")

    routes.Handle("/error", middlewareHandler.Handle([]rye.Handler{
        middlewareFirstHandler,
        errorHandler,
        homeHandler,
    })).Methods("GET")

    log.Infof("API server listening on %v", "localhost:8181")

    srv := &http.Server{
        Addr:    "localhost:8181",
        Handler: routes,
    }

    srv.ListenAndServe()
}

func homeHandler(rw http.ResponseWriter, r *http.Request) *rye.middlewares.Response {
    log.Infof("Home handler has fired!")

    fmt.Fprint(rw, "This is the home handler")
    return nil
}

func middlewareFirstHandler(rw http.ResponseWriter, r *http.Request) *rye.DetailedError {
    log.Infof("Middleware handler has fired!")
    return nil
}

func errorHandler(rw http.ResponseWriter, r *http.Request) *rye.DetailedError {
    log.Infof("Error handler has fired!")

    message := "This is the error handler"

    return &rye.DetailedError{
        StatusCode: http.StatusInternalServerError,
        Err:        errors.New(message),
    }
}
```

## API

### Config
This struct is configuration for the MWHandler. It holds references and config to dependencies such as the statsdClient and CORS related settings.
```go
type Config struct {
    Statter          statsd.Statter
    StatRate         float32
    CORSAllowOrigin  string
    CORSAllowMethods string
    CORSAllowHeaders string
}
```

### CORS
`rye` supports CORS! When instantiating `Config`, if you do NOT specify `CORSAllowOrigin` or `CORSAllowMethods` or `CORSAllowHeaders`, `rye` will fall back to DEFAULT values if it receives a request with the `Origin` header.

*Default* CORS Values:

**DEFAULT_CORS_ALLOW_METHODS**: "POST, GET, OPTIONS, PUT, DELETE"
**DEFAULT_CORS_ALLOW_HEADERS**: "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Access-Token"

If `CORSAllowOrigin` is NOT set in the config (and `Origin` header is passed), rye will set the `Access-Control-Allow-Origin` response header to whatever `Origin` was set to in the request.

In other words - you should probably properly instantiate `Config` with all CORS attributes in production.

### MWHandler
This struct is the primary handler container. It holds references to the statsd client.
```go
type MWHandler struct {
    Config Config
}
```
#### Constructor
```go
func NewMWHandler(statter statsd.Statter, statrate float32) *MWHandler
```
#### Handle
This method chains middleware handlers in order and returns a complete `http.Handler`.
```go
func (m *MWHandler) Handle(handlers []Handler) http.Handler
```

### DetailedError
This struct is for usage with the Handler type. This error adds StatusCode but fulfills the standard go Error interface through an `Error()` method.
```go
type DetailedError struct {
	Err        error
	StatusCode int
}
```

### Handler
This type is used to define an http handler that can be chained using the MWHandler.Handle method. The detailed error is from the **rye** package and has facilities to emit StatusCode.
```go
type Handler func(w http.ResponseWriter, r *http.Request) *DetailedError
```



## Test stuff
All interfacing with the project is done via `make`. Targets exist for all primary tasks such as:

- Testing: `make test` or `make testv` (for verbosity)
- Generate: `make generate` - this generates based on vendored libraries (from $GOPATH)
- All (test, build): `make all`
- .. and a few others. Run `make help` to see all available targets.
