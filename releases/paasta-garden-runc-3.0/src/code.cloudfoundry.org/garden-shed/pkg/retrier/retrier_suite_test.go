package retrier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRetrier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Retrier Suite")
}
