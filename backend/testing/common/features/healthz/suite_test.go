package healthz_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHealthzModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Suite common/features/healthz]")
}
