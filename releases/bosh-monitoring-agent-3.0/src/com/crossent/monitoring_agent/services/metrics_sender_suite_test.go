package services_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRuntimeSystemStats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RuntimeSystemStats Suite")
}

