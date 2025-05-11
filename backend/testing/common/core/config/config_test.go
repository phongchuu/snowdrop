package core_test

import (
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"internal.snowdrop/common/core"
)

var _ = Describe("[AppConfig]", func() {
	Describe("NewAppConfig", func() {
		It("should create a new app config", func() {
			appCfg, err := core.NewAppConfig(core.ModuleConfig{})
			Expect(err).NotTo(HaveOccurred())
			Expect(appCfg).NotTo(BeNil())
		})
	})

	Describe("GetDatabaseURL", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("APP_DB_USERNAME", "testuser")
			GinkgoT().Setenv("APP_DB_PASSWORD", "testpass")
			GinkgoT().Setenv("APP_DB_HOST", "localhost")
			GinkgoT().Setenv("APP_DB_NAME", "testdb")
			GinkgoT().Setenv("APP_DB_PORT", "5432")
		})

		It("should return the correct database URL", func() {
			appCfg, err := core.NewAppConfig(core.ModuleConfig{})

			Expect(err).NotTo(HaveOccurred())
			Expect(appCfg).NotTo(BeNil())
			Expect(appCfg.GetDatabaseURL()).To(Equal("postgresql://testuser:testpass@localhost:5432/testdb"))
		})
	})

	Describe("GetEmbedResourceFolder", func() {
		It("should return the correct embed resource folder", func() {
			var embedFS fstest.MapFS

			appCfg, err := core.NewAppConfig(core.ModuleConfig{
				EmbedResourcesFolder: embedFS,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(appCfg).NotTo(BeNil())
			Expect(appCfg.GetEmbedResourceFolder()).To(Equal(embedFS))
		})
	})
})
