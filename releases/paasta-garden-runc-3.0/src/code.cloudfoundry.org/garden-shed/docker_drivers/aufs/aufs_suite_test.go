package aufs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAufs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aufs Suite")
}
