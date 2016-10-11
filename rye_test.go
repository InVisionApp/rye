package rye

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/InVisionApp/rye/fakes/statsdfakes"
	"os"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

type statsInc struct {
	Name string
	Time int64
	StatRate float32
}

type statsTiming struct {
	Name string
	Time time.Duration
	StatRate float32
}

var _ = Describe("Rye", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder
		mwHandler	*MWHandler
		fakeStatter	*statsdfakes.FakeStatter
	)

	const (
		STATRATE float32 = 1
	)

	BeforeEach(func() {
		fakeStatter = &statsdfakes.FakeStatter{}
		ryeConfig := Config{fakeStatter,STATRATE}
		mwHandler = NewMWHandler(ryeConfig)

		response = httptest.NewRecorder()

		os.Unsetenv("RYE_TEST_HANDLER_PASS")
	})

	AfterEach(func() {
		os.Unsetenv("RYE_TEST_HANDLER_PASS")
	})

	Describe("NewMWHandler", func() {
		Context("when instantiating a mwhandler", func() {
			It("should have correct attributes", func() {
				ryeConfig := Config{fakeStatter,STATRATE}
				handler := NewMWHandler(ryeConfig)
				Expect(handler).NotTo(BeNil())
				Expect(handler.Config.Statter).To(Equal(fakeStatter))
				Expect(handler.Config.StatRate).To(Equal(STATRATE))
			})

			It("should have attributes with default values when passed an empty config", func() {
				handler := NewMWHandler(Config{})
				Expect(handler).NotTo(BeNil())
				Expect(handler.Config.Statter).To(BeNil())
				Expect(handler.Config.StatRate).To(Equal(float32(0.0)))
			})
		})
	})

	Describe("Handle", func() {
		Context("when adding a valid handler", func() {
			It("should return valid HandlerFunc", func() {

				var (
					actualInc statsInc
					actualTiming statsTiming
				)

				inc := make(chan statsInc)
				timing := make(chan statsTiming)

				fakeStatter.IncStub = func(name string, time int64, statrate float32) error {
					inc <- statsInc{name, time, statrate}
					close(inc)
					return nil
				}

				fakeStatter.TimingDurationStub = func(name string, time time.Duration, statrate float32) error {
					timing <- statsTiming{name, time, statrate}
					close(timing)
					return nil
				}

				h := mwHandler.Handle([]Handler{successHandler})
				h.ServeHTTP(response, request)

				actualInc = <-inc
				actualTiming = <-timing

				Expect(h).ToNot(BeNil())
				Expect(h).To(BeAssignableToTypeOf(func(http.ResponseWriter, *http.Request) {}))
				Expect(os.Getenv("RYE_TEST_HANDLER_PASS")).To(Equal("1"))
				Expect(actualInc.Name).To(Equal("handlers.successHandler.2xx"))
				Expect(actualInc.Time).To(Equal(int64(1)))
				Expect(actualInc.StatRate).To(Equal(float32(STATRATE)))
				Expect(actualTiming.Name).To(Equal("handlers.successHandler.runtime"))
				Expect(actualTiming.StatRate).To(Equal(float32(STATRATE)))
			})
		})

		Context("when adding an erroneous handler", func() {
			It("should interrupt handler chain and set a response status code", func() {

				var (
					actualInc statsInc
					actualTiming statsTiming
				)

				inc := make(chan statsInc)
				timing := make(chan statsTiming)

				fakeStatter.IncStub = func(name string, time int64, statrate float32) error {
					inc <- statsInc{name, time, statrate}
					close(inc)
					return nil
				}

				fakeStatter.TimingDurationStub = func(name string, time time.Duration, statrate float32) error {
					timing <- statsTiming{name, time, statrate}
					close(timing)
					return nil
				}

				h := mwHandler.Handle([]Handler{failureHandler})
				h.ServeHTTP(response, request)

				actualInc = <-inc
				actualTiming = <-timing

				Expect(h).ToNot(BeNil())
				Expect(h).To(BeAssignableToTypeOf(func(http.ResponseWriter, *http.Request) {}))
				Expect(response.Code).To(Equal(505))
				Expect(actualInc.Name).To(Equal("handlers.failureHandler.505"))
				Expect(actualInc.Time).To(Equal(int64(1)))
				Expect(actualInc.StatRate).To(Equal(float32(STATRATE)))
				Expect(actualTiming.Name).To(Equal("handlers.failureHandler.runtime"))
				Expect(actualTiming.StatRate).To(Equal(float32(STATRATE)))
			})
		})

		Context("when the statter is not set", func() {
			It("should not call Inc or TimingDuration", func() {

				ryeConfig := Config{}
				handler := NewMWHandler(ryeConfig)

				h := handler.Handle([]Handler{successHandler})
				h.ServeHTTP(response, request)

				Expect(fakeStatter.IncCallCount()).To(Equal(0))
				Expect(fakeStatter.TimingDurationCallCount()).To(Equal(0))
			})
		})
	})

	Describe("getFuncName", func() {
		It("should return the name of the function as a string", func() {
			funcName := getFuncName(testFunc)
			Expect(funcName).To(Equal("testFunc"))
		})
	})

})

func successHandler(rw http.ResponseWriter, r *http.Request) *DetailedError {
	os.Setenv("RYE_TEST_HANDLER_PASS", "1")
	return nil
}

func failureHandler(rw http.ResponseWriter, r *http.Request) *DetailedError {
	return &DetailedError{
		StatusCode: 505,
		Err:      fmt.Errorf("Foo"),
	}
}

func testFunc() {}
