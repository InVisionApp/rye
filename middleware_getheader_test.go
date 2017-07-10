package rye

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get Header Middleware", func() {
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

	Describe("getHeaderMiddleware", func() {
		Context("when a valid header is passed", func() {
			It("should return context with value", func() {
				headerName := "SpecialHeader"
				ctxKey := "special"
				request.Header.Add(headerName, "secret value")
				resp := NewMiddlewareGetHeader(headerName, ctxKey)(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Context).ToNot(BeNil())
				Expect(resp.Context.Value(ctxKey)).To(Equal("secret value"))
			})
		})

		Context("when no header is passed", func() {
			It("should have no value in context", func() {
				resp := NewMiddlewareGetHeader("something", "not there")(response, request)
				Expect(resp).To(BeNil())
			})
		})
	})
})
