package rye

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Static File Middleware", func() {

	var (
		request  *http.Request
		response *httptest.ResponseRecorder

		path     string
		testPath string
	)

	BeforeEach(func() {
		response = httptest.NewRecorder()
		request = &http.Request{}
		testPath, _ = os.Getwd()
	})

	Describe("handle", func() {
		Context("when a valid file is referenced", func() {
			It("should return a response", func() {
				path = "/static-examples/dist/index.html"
				url, _ := url.Parse("/thisstuff")
				request.URL = url
				resp := NewStaticFile(testPath+path)(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(200))

				body, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(body).To(ContainSubstring("Index.html"))
			})

			It("should return a Moved Permanently response", func() {
				path = "/static-examples/dist/index.html"
				url, _ := url.Parse("/thisstuff")
				request.URL = url
				resp := NewStaticFile("")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(301))
			})

			It("should return a File Not Found response", func() {
				path = "/static-examples/dist/index.html"
				url, _ := url.Parse("/thisstuff")
				request.URL = url
				resp := NewStaticFile(path)(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(404))
			})
		})
	})
})
