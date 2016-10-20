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

func NewMWHandler(config Config) *MWHandler {
	return &MWHandler{
		Config: config,
	}
}

type JSONStatus struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

type DetailedError struct {
	Err        error
	StatusCode int
}

// meet the Error interface
func (d *DetailedError) Error() string {
	return d.Err.Error()
}

//Handler Borrowed from http://laicos.com/writing-handsome-golang-middleware/
type Handler func(w http.ResponseWriter, r *http.Request) *DetailedError

func (m *MWHandler) Handle(handlers []Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func WriteJSONStatus(rw http.ResponseWriter, status, message string, statusCode int) {
	jsonData, _ := json.Marshal(&JSONStatus{
		Message: message,
		Status:  status,
	})

	WriteJSONResponse(rw, statusCode, jsonData)
}

func WriteJSONResponse(rw http.ResponseWriter, statusCode int, content []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(content)
}

func getFuncName(i interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	ns := strings.Split(fullName, ".")

	// when we get a method (not a raw function) it comes attached to whatever struct is in its
	// method receiver via a function closure, this is not precisely the same as that method itself
	// so the compiler appends "-fm" so the name of the closure does not conflict with the actual function
	// http://grokbase.com/t/gg/golang-nuts/153jyb5b7p/go-nuts-fm-suffix-in-function-name-what-does-it-mean#20150318ssinqqzrmhx2ep45wjkxsa4rua
	return strings.TrimSuffix(ns[len(ns)-1], ")-fm")
}
