package rye

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/InVisionApp/rye/fakes/statsdfakes"
)


var _ = Describe("Rye", func() {

	var (
		mwHandler	*MWHandler
		fakeStatter	*statsdfakes.FakeStatter
	)

	const (
		STATRATE float32 = 1
	)

	BeforeEach(func() {

		fakeStatter = &statsdfakes.FakeStatter{}
		mwHandler = NewMWHandler(fakeStatter,STATRATE)

	})


	Describe("NewMWHandler", func() {
		It("should return a non-nil MWHandler", func() {
			handler := NewMWHandler(nil, STATRATE)
			Expect(handler).NotTo(BeNil())
		})
	})


	Describe("getFuncName", func() {
		It("should return the name of the function as a string", func() {
			funcName := getFuncName(testFunc)
			Expect(funcName).To(Equal("testFunc"))
		})
	})

})

func testFunc() {}
