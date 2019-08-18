package csv_crm_upload

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCsvCrmUpload(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CsvCrmUpload Suite")
}
