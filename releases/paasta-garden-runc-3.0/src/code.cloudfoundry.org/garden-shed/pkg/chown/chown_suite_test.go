package chown_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestChown(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chown Suite")
}
