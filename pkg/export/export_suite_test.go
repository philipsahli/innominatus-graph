package export_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExportSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Export Suite")
}
