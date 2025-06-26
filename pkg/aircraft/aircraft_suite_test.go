package aircraft_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAircraft(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aircraft Suite")
}
