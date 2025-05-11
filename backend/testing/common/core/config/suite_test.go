package core_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Suite common/core/config]")
}
