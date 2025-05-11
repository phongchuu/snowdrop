package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuthModule(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Suite common/features/auth]")
}
