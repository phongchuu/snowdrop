package response_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResponseModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Module: framework/snowdrop/response")
}
