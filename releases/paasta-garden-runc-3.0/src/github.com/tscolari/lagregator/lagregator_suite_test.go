package lagregator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLagregator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Lagregator Suite")
}
