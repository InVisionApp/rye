package rye

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
)

//go:generate counterfeiter -o fakes/statsdfakes/fake_statter.go $GOPATH/src/github.com/cactus/go-statsd-client/statsd/client.go Statter
//go:generate perl -pi -e 's/$GOPATH\/src\///g' fakes/statsdfakes/fake_statter.go

// Middleware
type MWHandler struct {
	Config Config
}

type Config struct {
	Statter  statsd.Statter
	StatRate float32
}

type JSONStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Response struct is utilized by middlewares as a way to share state;
// ie. a middleware can return a *Response as a way to indicate
// that further middleware execution should stop (without an error) or return a
// a hard error by setting `Err` + `StatusCode`.
type Response struct {
	Err           error
	StatusCode    int
	StopExecution bool
}

// Meet the Error interface
func (r *Response) Error() string {
	return r.Err.Error()
}

//Handler Borrowed from http://laicos.com/writing-handsome-golang-middleware/
type Handler func(w http.ResponseWriter, r *http.Request) *Response

// Constructor for new instantiating new rye instances
func NewMWHandler(config Config) *MWHandler {
	return &MWHandler{
		Config: config,
	}
}

func (m *MWHandler) Handle(handlers []Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			var resp *Response

			// Record handler runtime
			func() {
				statusCode := "2xx"
				startTime := time.Now()

				if resp = handler(w, r); resp != nil {
					if resp.StopExecution {
						return
					}

					// Middleware did something funky - returned a *Response
					// but did not set an error;
					if resp.Err == nil {
						resp.Err = errors.New("Problem with middleware; neither Err or StopExecution is set")
						resp.StatusCode = http.StatusInternalServerError
					}

					if m.Config.Statter != nil && resp.StatusCode >= 500 {
						go m.Config.Statter.Inc("errors", 1, m.Config.StatRate)
					}

					statusCode = strconv.Itoa(resp.StatusCode)
					WriteJSONStatus(w, "error", resp.Error(), resp.StatusCode)
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
			if resp != nil {
				return
			}
		}
	})
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
