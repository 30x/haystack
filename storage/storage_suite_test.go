package storage_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStorageSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Test")
}
