package garden_integration_tests_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Debug struct {
	MemStats runtime.MemStats `json:"memstats"`
}

func loadDebug() Debug {
	response, err := http.Get(fmt.Sprintf("http://%s:%s/debug/vars", gardenHost, gardenDebugPort))
	Expect(err).NotTo(HaveOccurred())
	defer response.Body.Close()

	Expect(err).NotTo(HaveOccurred())
	debug := &Debug{}
	body, err := ioutil.ReadAll(response.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(body, debug)).To(Succeed())

	return *debug
}

var _ = Describe("Debug", func() {
	Describe("Memory", func() {
		It("should have non-zero allocated memory", func() {
			debug := loadDebug()
			Expect(debug.MemStats.Alloc).NotTo(BeZero())
		})
	})
})
