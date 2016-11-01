package middleware

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base tests", func() {
	Describe("Error()", func() {
		Context("when an error is set on Response struct", func() {
			It("should return a string if you call Error()", func() {
				resp := &Response{
					Err: errors.New("some error"),
				}

				Expect(resp.Error()).To(Equal("some error"))
			})
		})
	})
})
