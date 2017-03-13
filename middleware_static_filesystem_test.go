package rye

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
				path = "/static-examples/dist/"

				request, _ := http.NewRequest("GET", "/dist/test.html", nil)

				resp := NewStaticFilesystem(testPath+path, "/dist/")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(200))

				body, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(body).To(ContainSubstring("Test.html"))
			})

			It("should return Index.html when request is just path", func() {
				path = "/static-examples/dist/"

				request, _ := http.NewRequest("GET", "/dist/", nil)

				resp := NewStaticFilesystem(testPath+path, "/dist/")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(200))

				body, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(body).To(ContainSubstring("Index.html"))
			})

			It("should return Index.html when strip prefix is empty", func() {
				path = "/static-examples/dist/"

				request, _ := http.NewRequest("GET", "/", nil)

				resp := NewStaticFilesystem(testPath+path, "")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(200))

				body, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(body).To(ContainSubstring("Index.html"))
			})

			It("should return Index.html when strip prefix is empty", func() {
				path = "/static-examples/dist/"

				request, _ := http.NewRequest("GET", "/ASDads.HTML", nil)

				resp := NewStaticFilesystem(testPath+path, "")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(404))
			})

			It("should return test.css on subpath", func() {
				path = "/static-examples/dist/"

				request, _ := http.NewRequest("GET", "/styles/test.css", nil)

				resp := NewStaticFilesystem(testPath+path, "")(response, request)
				Expect(resp).To(BeNil())
				Expect(response).ToNot(BeNil())
				Expect(response.Code).To(Equal(200))

				body, err := ioutil.ReadAll(response.Body)
				Expect(err).To(BeNil())
				Expect(body).To(ContainSubstring("test.css"))
			})
		})
	})
})
