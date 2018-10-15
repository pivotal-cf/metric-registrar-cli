package registrations_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRegistrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registrations Suite")
}
