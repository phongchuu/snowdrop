package auth_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAuthRoutes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "[Suite Auth]")
}
