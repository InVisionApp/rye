package rye

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CIDR Middleware", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder

		cidr1, cidr2, ip1, ip2, ip3 string
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
		request = &http.Request{}
		cidr1 = "10.0.0.0/24"
		cidr2 = "127.0.0.0/24"
		ip1 = "10.0.0.1:22"
		ip2 = "127.0.0.1:22"
		ip3 = "192.0.0.1:22"
	})

	Describe("handle", func() {
		Context("when a valid IP is used", func() {
			It("should return nil", func() {
				request.RemoteAddr = ip1
				resp := NewMiddlewareCIDR([]string{cidr1, cidr2})(response, request)
				Expect(resp).To(BeNil())
			})

			It("should return nil", func() {
				request.RemoteAddr = ip2
				resp := NewMiddlewareCIDR([]string{cidr1, cidr2})(response, request)
				Expect(resp).To(BeNil())
			})
		})

		Context("when an invalid IP is used", func() {
			It("should return an error", func() {
				request.RemoteAddr = ip3
				resp := NewMiddlewareCIDR([]string{cidr1, cidr2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("not authorized"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when no IP exists", func() {
			It("should return an error", func() {
				resp := NewMiddlewareCIDR([]string{cidr1, cidr2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("Remote address error"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when an unrecognizable IP is used", func() {
			It("should return an error", func() {
				request.RemoteAddr = "blah:80"
				resp := NewMiddlewareCIDR([]string{cidr1, cidr2})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("Error validating IP address: Unable to parse IP"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when an unrecognizable CIDR is used", func() {
			It("should return an error", func() {
				request.RemoteAddr = ip1
				resp := NewMiddlewareCIDR([]string{"blah"})(response, request)
				Expect(resp).ToNot(BeNil())
				Expect(resp.Err).To(HaveOccurred())
				Expect(resp.Error()).To(ContainSubstring("Error validating IP address: invalid CIDR address"))
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

	})
})
