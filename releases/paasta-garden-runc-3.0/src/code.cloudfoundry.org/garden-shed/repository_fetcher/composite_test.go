package repository_fetcher_test

import (
	"net/url"

	. "code.cloudfoundry.org/garden-shed/repository_fetcher"
	fakes "code.cloudfoundry.org/garden-shed/repository_fetcher/repository_fetcherfakes"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompositeFetcher", func() {
	var (
		logger            *lagertest.TestLogger
		fakeLocalFetcher  *fakes.FakeRepositoryFetcher
		fakeRemoteFetcher *fakes.FakeRepositoryFetcher
		factory           *CompositeFetcher
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		fakeLocalFetcher = new(fakes.FakeRepositoryFetcher)
		fakeRemoteFetcher = new(fakes.FakeRepositoryFetcher)

		factory = &CompositeFetcher{
			LocalFetcher:  fakeLocalFetcher,
			RemoteFetcher: fakeRemoteFetcher,
		}
	})

	Context("when the URL does not contain a scheme", func() {
		It("delegates .Fetch to the local fetcher", func() {
			factory.Fetch(logger, &url.URL{Path: "cake"}, "", "", 24)
			Expect(fakeLocalFetcher.FetchCallCount()).To(Equal(1))
			Expect(fakeRemoteFetcher.FetchCallCount()).To(Equal(0))
		})

		It("delegates .FetchID to the local fetcher", func() {
			factory.FetchID(logger, &url.URL{Path: "cake"})
			Expect(fakeLocalFetcher.FetchIDCallCount()).To(Equal(1))
			Expect(fakeRemoteFetcher.FetchIDCallCount()).To(Equal(0))
		})
	})

	Context("when the scheme is docker://", func() {
		It("delegates .Fetch to the remote fetcher", func() {
			factory.Fetch(logger, &url.URL{Scheme: "docker", Path: "cake"}, "", "", 24)
			Expect(fakeRemoteFetcher.FetchCallCount()).To(Equal(1))
			Expect(fakeLocalFetcher.FetchCallCount()).To(Equal(0))
		})

		It("delegates .FetchID to the remote fetcher", func() {
			factory.FetchID(logger, &url.URL{Scheme: "docker", Path: "cake"})
			Expect(fakeRemoteFetcher.FetchIDCallCount()).To(Equal(1))
			Expect(fakeLocalFetcher.FetchIDCallCount()).To(Equal(0))
		})
	})
})
