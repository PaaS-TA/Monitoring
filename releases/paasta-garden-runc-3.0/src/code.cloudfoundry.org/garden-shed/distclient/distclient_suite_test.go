package distclient_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDistclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distclient Suite")
}
