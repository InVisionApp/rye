package rye

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessToken Middleware", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder

		tokenHeaderName = "at-hname"
		token1, token2  string
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
		request = &http.Request{
			Header: map[string][]string{},
		}

		token1 = "test1"
		token2 = "test2"
	})

	Describe("handle", func() {
		Context("when a valid token is used", func() {
			It("should return nil", func() {
				request.Header.Add(tokenHeaderName, token1)
				resp := NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})(response, request)
				Expect(resp).To(BeNil())
			})

			It("should return nil", func() {
				request.Header.Add(tokenHeaderName, token2)
				resp := NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when an invalid token is used", func() {
			It("should return an error", func() {
				request.Header.Add(tokenHeaderName, "blah")
				resp := NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("invalid access token"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when no token header exists", func() {
			It("should return an error", func() {
				resp := NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when token header is blank", func() {
			It("should return an error", func() {
				request.Header.Add(tokenHeaderName, "")
				resp := NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
