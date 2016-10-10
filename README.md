# rye
A simple library to support http services. Currently, **rye** provides a middleware handler which can be used to chain http handlers together while providing statsd timing and status code for use with DataDog or other logging aggregators.

## Setup
In order to use **rye**, you should vendor it and the **statsd** client within your project.

```
govendor fetch github.com/cactus/go-statsd-client/statsd
govendor fetch github.com/InVisionApp/rye
```
## Example

Begin by importing the required libraries:

```go
import (
    "github.com/cactus/go-statsd-client/statsd"
    "github.com/InVisionApp/rye"
)
```

Create a statsd client and a middleware handler. The purpose of the Handler is to keep a reference to the statsd client for gathering metrics and to provide an interface for chaining http handlers.
```go
middlewareHandler := rye.NewMWHandler(statsdClient, DEFAULT_STATSD_RATE)
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
import (
    "github.com/cactus/go-statsd-client/statsd"
    "github.com/InVisionApp/rye"
)

// Config vars omitted (StatsDAddress, StatsDPrefix, flushInterval, DEFAULT_STATSD_RATE, etc

func main() {

    statsdClient, err := statsd.NewBufferedClient(StatsDAddress, StatsDPrefix, flushInterval, 0)
	if err != nil {
		log.Fatalf("Unable to instantiate statsd client: %v", err.Error())
	}

    middlewareHandler := rye.NewMWHandler(statsdClient, DEFAULT_STATSD_RATE)

    routes := mux.NewRouter().StrictSlash(true)

    routes.Handle("/", middlewareHandler.Handle([]rye.Handler{
        middlewareFirstHandler,
        homeHandler,
    })).Methods("GET")

    routes.Handle("/error", middlewareHandler.Handle([]rye.Handler{
        middlewareFirstHandler,
        homeHandler,
        errorHandler
    })).Methods("GET")

    log.Infof("API server listening on %v", ListenAddress)

    srv := &http.Server{
        Addr:         ListenAddress,
        Handler:      routes,
        ReadTimeout:  time.Duration(ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(WriteTimeout) * time.Second,
    }

    srv.ListenAndServe()
}

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

## API

### MWHandler
This struct is the primary handler container. It holds references to the statsd client.
```go
type MWHandler struct {
	Statter  statsd.Statter
	StatRate float32
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



## Build stuff
All interfacing with the project is done via `make`. Targets exist for all primary tasks such as:

- Building: `make build` (or `make build/linux`, `make build/darwin`)
- Testing: `make test` (or `make test/unit`, `make test/integration`)
- All (test, build): `make all`
- .. and a few others. Run `make help` to see all available targets.