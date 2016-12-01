package rye

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Logger Middleware", func() {

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

	Describe("MiddlewareRouteLogger", func() {
		Context("when the route logger is called", func() {
			It("should return nil", func() {
				resp := MiddlewareRouteLogger()(response, request)
				Expect(resp).To(BeNil())
			})
		})
	})
})
