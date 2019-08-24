package sender

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSender(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sender Suite")
}
