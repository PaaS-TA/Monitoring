package distclient_test

import (
	"code.cloudfoundry.org/garden-shed/distclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InsecureRegistryList", func() {
	Context("when a list of secure repositories is provided", func() {
		Context("and the requested endpoint is in the list", func() {
			It("returns that the registry is insecure", func() {
				provider := distclient.InsecureRegistryList{"insecure1", "insecure2"}
				Expect(provider.AllowInsecure("insecure1")).To(Equal(true))
			})

			Context("and the list is using an IP", func() {
				It("returns that the registry is insecure", func() {
					provider := distclient.InsecureRegistryList{"100.100.100.0/24", "103.100.100.15"}
					Expect(provider.AllowInsecure("103.100.100.15")).To(Equal(true))
				})
			})

			Context("and the list is using CIDR addresses", func() {
				It("returns that the registry is insecure", func() {
					provider := distclient.InsecureRegistryList{"100.100.100.0/24", "103.100.100.0/24"}
					Expect(provider.AllowInsecure("100.100.100.155")).To(Equal(true))
				})

				It("returns that the registry is insecure, even if the host contains a port", func() {
					provider := distclient.InsecureRegistryList{"100.100.100.0/24", "103.100.100.0/24"}
					Expect(provider.AllowInsecure("100.100.100.155:5000")).To(Equal(true))
				})
			})

			Context("and the list is using CIDR addresses and hostnames", func() {
				It("returns that the registry is insecure", func() {
					provider := distclient.InsecureRegistryList{"100.100.100.0/24", "103.100.100.0/24", "hostname1"}
					Expect(provider.AllowInsecure("hostname1")).To(Equal(true))
				})
			})
		})

		Context("and the requested endpoint is not in the list", func() {
			It("assumes the registry is secure", func() {
				provider := distclient.InsecureRegistryList{"insecure1", "insecure2"}
				Expect(provider.AllowInsecure("the-registry-host:44")).To(Equal(false))
			})

			Context("and the list is using CIDR addresses", func() {
				It("returns that the registry is secure", func() {
					provider := distclient.InsecureRegistryList{"100.100.100.0/24", "103.100.100.0/24"}
					Expect(provider.AllowInsecure("100.100.95.155")).To(Equal(false))
				})
			})
		})
	})
})
