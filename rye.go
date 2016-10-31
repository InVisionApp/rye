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

	"github.com/InVisionApp/rye/middlewares"
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

//Handler Borrowed from http://laicos.com/writing-handsome-golang-middleware/
type Handler func(w http.ResponseWriter, r *http.Request) *middlewares.Response

// Constructor for new instantiating new rye instances
func NewMWHandler(config Config) *MWHandler {
	return &MWHandler{
		Config: config,
	}
}

func (m *MWHandler) Handle(handlers []Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			var resp *middlewares.Response

			// Record handler runtime
			func() {
				statusCode := "2xx"
				startTime := time.Now()

				if resp = handler(w, r); resp != nil {
					if resp.StopExecution {
						return
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
