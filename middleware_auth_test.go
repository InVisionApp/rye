package rye

import (
	"net/http"
	"net/http/httptest"

	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const AUTH_HEADER_NAME = "Authorization"

var _ = Describe("Auth Middleware", func() {
	var (
		request  *http.Request
		response *httptest.ResponseRecorder

		testHandler func(http.ResponseWriter, *http.Request) *Response
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
	})

	Context("auth", func() {
		var (
			fakeAuth *recorder
		)

		BeforeEach(func() {
			fakeAuth = &recorder{}

			testHandler = NewMiddlewareAuth(fakeAuth.authFunc)
			request = &http.Request{
				Header: map[string][]string{},
			}
		})

		It("passes the header to the auth func", func() {
			testAuth := "foobar"
			request.Header.Add(AUTH_HEADER_NAME, testAuth)
			resp := testHandler(response, request)

			Expect(resp).To(BeNil())
			Expect(fakeAuth.header).To(Equal(testAuth))
		})

		Context("when no header is found", func() {
			It("errors", func() {
				resp := testHandler(response, request)

				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Error()).To(ContainSubstring("no authentication"))
			})
		})
	})

	Context("Basic Auth", func() {
		var (
			username = "user1"
			pass     = "mypass"
		)

		BeforeEach(func() {
			testHandler = NewMiddlewareAuth(NewBasicAuthFunc(map[string]string{
				username: pass,
			}))

			request = &http.Request{
				Header: map[string][]string{},
			}
		})

		It("validates the password", func() {
			request.SetBasicAuth(username, pass)
			resp := testHandler(response, request)

			Expect(resp.Err).To(BeNil())
		})

		It("adds the username to context", func() {
			request.SetBasicAuth(username, pass)
			resp := testHandler(response, request)

			Expect(resp.Err).To(BeNil())

			ctxUname := resp.Context.Value(AUTH_USERNAME_KEY)
			uname, ok := ctxUname.(string)
			Expect(ok).To(BeTrue())
			Expect(uname).To(Equal(username))
		})

		It("preserves the request context", func() {

		})

		It("errors if username unknown", func() {
			request.SetBasicAuth("noname", pass)
			resp := testHandler(response, request)

			Expect(resp.Err).ToNot(BeNil())
			Expect(resp.Err.Error()).To(ContainSubstring("invalid auth"))
		})

		It("errors if password wrong", func() {
			request.SetBasicAuth(username, "wrong")
			resp := testHandler(response, request)

			Expect(resp.Err).ToNot(BeNil())
			Expect(resp.Err.Error()).To(ContainSubstring("invalid auth"))
		})

		Context("parseBasicAuth", func() {
			It("errors if header not basic", func() {
				request.Header.Add(AUTH_HEADER_NAME, "wrong")
				resp := testHandler(response, request)

				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Error()).To(ContainSubstring("invalid auth"))
			})

			It("errors if header not base64", func() {
				request.Header.Add(AUTH_HEADER_NAME, "Basic ------")
				resp := testHandler(response, request)

				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Error()).To(ContainSubstring("invalid auth"))
			})

			It("errors if header wrong format", func() {
				request.Header.Add(AUTH_HEADER_NAME, "Basic YXNkZgo=") // asdf no `:`
				resp := testHandler(response, request)

				Expect(resp.Err).ToNot(BeNil())
				Expect(resp.Err.Error()).To(ContainSubstring("invalid auth"))
			})
		})
	})
})

type recorder struct {
	header string
}

func (r *recorder) authFunc(ctx context.Context, s string) *Response {
	r.header = s
	return nil
}
