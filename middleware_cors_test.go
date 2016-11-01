package rye

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CORS Middleware", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
		request = &http.Request{
			Header: make(map[string][]string, 0),
		}
	})

	Describe("handle", func() {
		Context("when origin header is not set", func() {
			It("should return nil", func() {
				resp := MiddlewareCORS()(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when origin header is set", func() {
			Context("and CORS was instantiated with params", func() {
				var (
					testOrigin  = "*.invisionapp.com"
					testHeaders = "TestHeader"
					testMethods = "GET, POST, TESTMETHOD"
				)

				It("should set all CORS headers from params", func() {
					request.Header.Add("Origin", "*.invisionapp.com")
					resp := NewMiddlewareCORS(testOrigin, testMethods, testHeaders)(response, request)

					Expect(resp).To(BeNil())
					Expect(response.Header().Get("Access-Control-Allow-Origin")).To(Equal(testOrigin))
					Expect(response.Header().Get("Access-Control-Allow-Methods")).To(Equal(testMethods))
					Expect(response.Header().Get("Access-Control-Allow-Headers")).To(Equal(testHeaders))
				})
			})

			Context("and CORS was instantiated with defaults", func() {
				It("should set all CORS headers using defaults", func() {
					request.Header.Add("Origin", "*.invisionapp.com")
					resp := MiddlewareCORS()(response, request)

					Expect(resp).To(BeNil())
					Expect(response.Header().Get("Access-Control-Allow-Origin")).To(Equal(DEFAULT_CORS_ALLOW_ORIGIN))
					Expect(response.Header().Get("Access-Control-Allow-Methods")).To(Equal(DEFAULT_CORS_ALLOW_METHODS))
					Expect(response.Header().Get("Access-Control-Allow-Headers")).To(Equal(DEFAULT_CORS_ALLOW_HEADERS))
				})
			})

			Context("and we got a preflight request (OPTIONS)", func() {
				It("should return a response with StopExecution", func() {
					request.Method = "OPTIONS"
					request.Header.Add("Origin", "*.invisionapp.com")
					resp := MiddlewareCORS()(response, request)

					Expect(resp).ToNot(BeNil())
					Expect(resp.StopExecution).To(BeTrue())
				})
			})
		})
	})
})
