package zetachain_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestZetachain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zetachain Suite")
}
