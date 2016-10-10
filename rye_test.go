package rye

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/InVisionApp/rye/fakes/statsdfakes"
	"os"
	"fmt"
	"net/http"
	"net/http/httptest"
)


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
		mwHandler = NewMWHandler(fakeStatter,STATRATE)

		response = httptest.NewRecorder()

		os.Unsetenv("RYE_TEST_HANDLER_PASS")
	})

	AfterEach(func() {
		os.Unsetenv("RYE_TEST_HANDLER_PASS")
	})

	Describe("NewMWHandler", func() {
		Context("when instantiating a mwhandler", func() {
			It("should have correct attributes", func() {
				handler := NewMWHandler(fakeStatter, STATRATE)
				Expect(handler).NotTo(BeNil())
				Expect(handler.Statter).To(Equal(fakeStatter))
				Expect(handler.StatRate).To(Equal(STATRATE))
			})
		})
	})

	Describe("Handle", func() {
		Context("when adding a valid handler", func() {
			It("should return valid HandlerFunc", func() {
				successHandler := func(rw http.ResponseWriter, r *http.Request) *DetailedError {
					os.Setenv("RYE_TEST_HANDLER_PASS", "1")
					return nil
				}

				h := mwHandler.Handle([]Handler{successHandler})
				h.ServeHTTP(response, request)

				Expect(h).ToNot(BeNil())
				Expect(h).To(BeAssignableToTypeOf(func(http.ResponseWriter, *http.Request) {}))
				Expect(os.Getenv("RYE_TEST_HANDLER_PASS")).To(Equal("1"))
			})
		})

		Context("when adding an erroneous handler", func() {
			It("should interrupt handler chain and set a response status code", func() {
				failureHandler := func(rw http.ResponseWriter, r *http.Request) *DetailedError {
					return &DetailedError{
						StatusCode: 505,
						Err:      fmt.Errorf("Foo"),
					}
				}

				h := mwHandler.Handle([]Handler{failureHandler})
				h.ServeHTTP(response, request)

				Expect(h).ToNot(BeNil())
				Expect(h).To(BeAssignableToTypeOf(func(http.ResponseWriter, *http.Request) {}))
				Expect(response.Code).To(Equal(505))
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

func testFunc() {}
