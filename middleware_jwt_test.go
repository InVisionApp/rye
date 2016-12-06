package rye

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT Middleware", func() {

	var (
		request       *http.Request
		response      *httptest.ResponseRecorder
		shared_secret = "secret"
		hs256_jwt     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ"
		rs256_jwt     = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2p3dC1pZHAuZXhhbXBsZS5jb20iLCJzdWIiOiJtYWlsdG86bWlrZUBleGFtcGxlLmNvbSIsIm5iZiI6MTQ3ODIwMTkxNiwiZXhwIjoxNDc4MjA1NTE2LCJpYXQiOjE0NzgyMDE5MTYsImp0aSI6ImlkMTIzNDU2IiwidHlwIjoiaHR0cHM6Ly9leGFtcGxlLmNvbS9yZWdpc3RlciJ9.B9zAEk_zm_Hz3cn8QLtAZNizZtHlZ0ENQ0nn5Jl734cYKO6Rn2JJct24u3UPXl01atIre2Z8oKIs9gpePpBsvR50Z-gCFtTGM_5dTPw45H4hY4KkvjP9JvnYGz4V4DeQDTZz-HUByKHSbKNm4pCmhGLcuF2SBwBmj-xOoy4eCc4Zf77fSXz9ctwv3FHCteXQnXD6M2m243fVkPiWq7qaE0Z0CfR0vRHjUcbA2qWVAM1kuOUSIIqhc0hZ5sVIW1UYZ4XHJ7unXG_SRRT6sYEE3FRRKhURutRRkhLtMMF14TpcQdZC0UjkcJVVMnQR0HQiDG-L7TModfNRhBO5PpjDng"
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
		request = &http.Request{
			Header: make(map[string][]string, 0),
		}
	})

	Describe("handle", func() {
		Context("when a valid token is passed", func() {
			It("should return nil", func() {
				request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", hs256_jwt))
				resp := NewMiddlewareJWT(shared_secret)(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Context).ToNot(BeNil())
				Expect(resp.Context.Value(CONTEXT_JWT)).To(Equal(hs256_jwt))
			})
		})

		Context("when no token is passed", func() {
			It("should return an error", func() {
				resp := NewMiddlewareJWT(shared_secret)(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Error()).To(ContainSubstring("JWT token must be passed"))
			})
		})

		Context("when an invalid token is passed", func() {
			It("should return an error", func() {
				request.Header.Add("Authorization", "Bearer foo")
				resp := NewMiddlewareJWT(shared_secret)(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Error()).To(ContainSubstring("invalid"))
			})
		})

		Context("when a token with an incorrectly signed signature is passed", func() {
			It("should return an error", func() {
				request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rs256_jwt))
				resp := NewMiddlewareJWT(shared_secret)(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Error()).To(ContainSubstring("signing method"))
			})
		})
	})
})
