package rye

import (
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
)

const (
	DEFAULT_CORS_ALLOW_METHODS = "POST, GET, OPTIONS, PUT, DELETE"
	DEFAULT_CORS_ALLOW_HEADERS = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Access-Token"
)

//go:generate counterfeiter -o fakes/statsdfakes/fake_statter.go $GOPATH/src/github.com/cactus/go-statsd-client/statsd/client.go Statter
//go:generate perl -pi -e 's/$GOPATH\/src\///g' fakes/statsdfakes/fake_statter.go

// Middleware
type MWHandler struct {
	Config Config
}

type Config struct {
	Statter          statsd.Statter
	StatRate         float32
	CORSAllowOrigin  string
	CORSAllowMethods string
	CORSAllowHeaders string
}

type JSONStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type DetailedError struct {
	Err        error
	StatusCode int
}

//Handler Borrowed from http://laicos.com/writing-handsome-golang-middleware/
type Handler func(w http.ResponseWriter, r *http.Request) *DetailedError

// Constructor for new instantiating new rye instances
func NewMWHandler(config Config) *MWHandler {
	return &MWHandler{
		Config: config,
	}
}

// Meet the Error interface
func (d *DetailedError) Error() string {
	return d.Err.Error()
}

func (m *MWHandler) Handle(handlers []Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handleCORS(w, r)

		// No need to go any further if preflight OPTIONS request
		if r.Method == "OPTIONS" {
			return
		}

		for _, handler := range handlers {
			var err *DetailedError

			// Record handler runtime
			func() {
				statusCode := "2xx"
				startTime := time.Now()

				if err = handler(w, r); err != nil {
					statusCode = strconv.Itoa(err.StatusCode)
					WriteJSONStatus(w, "error", err.Error(), err.StatusCode)
				}

				handlerName := getFuncName(handler)

				if m.Config.Statter != nil {
					// Record runtime metric
					go m.Config.Statter.TimingDuration(
						"handlers."+handlerName+".runtime",
						time.Since(startTime), // delta
						m.Config.StatRate,
					)

					// Record status code metric (default 2xx)
					go m.Config.Statter.Inc(
						"handlers."+handlerName+"."+statusCode,
						1,
						m.Config.StatRate,
					)
				}
			}()

			// stop executing rest of the handlers if we encounter an error
			if err != nil {
				return
			}
		}
	})
}

// If `Origin` header gets passed, add required response headers for CORS support.
// Return bool if `Origin` header was detected.
func (m *MWHandler) handleCORS(rw http.ResponseWriter, req *http.Request) bool {
	origin := req.Header.Get("Origin")

	if origin == "" {
		return false
	}

	// if Config.CORSAllowOrigin is set - use that, otherwise fall back to provided origin
	if m.Config.CORSAllowOrigin != "" {
		origin = m.Config.CORSAllowOrigin
	}

	// if Config.CORSAllowMethods is set - use that, otherwise fall back to default allowed methods
	allowMethods := DEFAULT_CORS_ALLOW_METHODS

	if m.Config.CORSAllowMethods != "" {
		allowMethods = m.Config.CORSAllowMethods
	}

	// if Config.AllowHeaders is set - use that, otherwise fall back to default allowed headers
	allowHeaders := DEFAULT_CORS_ALLOW_HEADERS

	if m.Config.CORSAllowHeaders != "" {
		allowHeaders = m.Config.CORSAllowHeaders
	}

	rw.Header().Set("Access-Control-Allow-Origin", origin)
	rw.Header().Set("Access-Control-Allow-Methods", allowMethods)
	rw.Header().Set("Access-Control-Allow-Headers", allowHeaders)

	return true
}

// Wrapper for WriteJSONResponse that returns a marshalled JSONStatus blob
func WriteJSONStatus(rw http.ResponseWriter, status, message string, statusCode int) {
	jsonData, _ := json.Marshal(&JSONStatus{
		Message: message,
		Status:  status,
	})

	WriteJSONResponse(rw, statusCode, jsonData)
}

// Write data and status code to rw
func WriteJSONResponse(rw http.ResponseWriter, statusCode int, content []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(content)
}

// Programmatically determine given function name (and perform string cleanup)
func getFuncName(i interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	ns := strings.Split(fullName, ".")

	// when we get a method (not a raw function) it comes attached to whatever struct is in its
	// method receiver via a function closure, this is not precisely the same as that method itself
	// so the compiler appends "-fm" so the name of the closure does not conflict with the actual function
	// http://grokbase.com/t/gg/golang-nuts/153jyb5b7p/go-nuts-fm-suffix-in-function-name-what-does-it-mean#20150318ssinqqzrmhx2ep45wjkxsa4rua
	return strings.TrimSuffix(ns[len(ns)-1], ")-fm")
}
