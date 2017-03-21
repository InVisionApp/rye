package rye

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccessToken Middleware", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder

		testHandler func(http.ResponseWriter, *http.Request) *Response

		token1, token2 string
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()

		token1 = "test1"
		token2 = "test2"
	})

	Context("header token", func() {
		var (
			tokenHeaderName = "at-hname"
		)

		BeforeEach(func() {
			testHandler = NewMiddlewareAccessToken(tokenHeaderName, []string{token1, token2})
			request = &http.Request{
				Header: map[string][]string{},
			}
		})

		Context("when a valid token is used", func() {
			It("should return nil", func() {
				request.Header.Add(tokenHeaderName, token1)
				resp := testHandler(response, request)
				Expect(resp).To(BeNil())
			})

			It("should return nil", func() {
				request.Header.Add(tokenHeaderName, token2)
				resp := testHandler(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when an invalid token is used", func() {
			It("should return an error", func() {
				request.Header.Add(tokenHeaderName, "blah")
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("invalid access token"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when no token header exists", func() {
			It("should return an error", func() {
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when token header is blank", func() {
			It("should return an error", func() {
				request.Header.Add(tokenHeaderName, "")
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Context("query param token", func() {
		var (
			qParamName string
			qParams    string
		)

		BeforeEach(func() {
			qParamName = "token"
			testHandler = NewMiddlewareAccessQueryToken(qParamName, []string{token1, token2})
		})

		JustBeforeEach(func() {
			u, err := url.Parse(fmt.Sprintf("http://doesntmatter.io/blah?%s", qParams))
			Expect(err).ToNot(HaveOccurred())

			request = &http.Request{
				URL: u,
			}
		})

		Context("when a valid token is used", func() {
			BeforeEach(func() {
				qParams = fmt.Sprintf("%s=%s", qParamName, token1)
			})

			It("should return nil", func() {
				resp := testHandler(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when the other valid token is used", func() {
			BeforeEach(func() {
				qParams = fmt.Sprintf("%s=%s", qParamName, token2)
			})

			It("should return nil", func() {
				resp := testHandler(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when an invalid token is used", func() {
			BeforeEach(func() {
				qParams = fmt.Sprintf("%s=blah", qParamName)
			})

			It("should return an error", func() {
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("invalid access token"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when no token param exists", func() {
			BeforeEach(func() {
				qParams = "something=else"
			})

			It("should return an error", func() {
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when token param is blank", func() {
			BeforeEach(func() {
				qParams = fmt.Sprintf("%s=''", qParamName)
			})

			It("should return an error", func() {
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("invalid access token"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when no query params", func() {
			JustBeforeEach(func() {
				u, err := url.Parse("http://doesntmatter.io/blah")
				Expect(err).ToNot(HaveOccurred())

				request = &http.Request{
					URL: u,
				}
			})

			It("should return an error", func() {
				resp := testHandler(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("No access token found"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

	})
})
