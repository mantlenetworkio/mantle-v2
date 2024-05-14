package bls_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBls(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bls Suite")
}
