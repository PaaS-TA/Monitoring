package repository_fetcher_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"sync"

	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/runconfig"

	"code.cloudfoundry.org/garden-shed/distclient"
	"code.cloudfoundry.org/garden-shed/distclient/fake_distclient"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/layercake/fake_cake"
	"code.cloudfoundry.org/garden-shed/repository_fetcher"
	fakes "code.cloudfoundry.org/garden-shed/repository_fetcher/repository_fetcherfakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/distribution/digest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetching from a Remote repo", func() {
	var (
		logger       *lagertest.TestLogger
		fakeDialer   *fakes.FakeDialer
		fakeConn     *fake_distclient.FakeConn
		fakeCake     *fake_cake.FakeCake
		fakeVerifier *fakes.FakeVerifier

		remote *repository_fetcher.Remote

		manifests                 map[string]*distclient.Manifest
		blobs                     map[digest.Digest]string
		existingLayers            map[string]bool
		defaultDockerRegistryHost string
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		defaultDockerRegistryHost = "registry-1.docker.io"
	})

	JustBeforeEach(func() {
		existingLayers = map[string]bool{}

		manifests = map[string]*distclient.Manifest{
			"latest": &distclient.Manifest{
				Layers: []distclient.Layer{
					{},
				},
			},
			"some-tag": &distclient.Manifest{
				Layers: []distclient.Layer{
					{
						BlobSum:        "abc-def",
						StrongID:       "sha256:abc-id",
						ParentStrongID: "sha256:abc-parent-id",
						Image: image.Image{
							Config: &runconfig.Config{
								Env:     []string{"a", "b"},
								Volumes: map[string]struct{}{"vol1": struct{}{}},
							},
							Size: 1,
						},
					},
					{
						BlobSum:  "ghj-klm",
						StrongID: "sha256:ghj-id",
					},
					{
						BlobSum:  "klm-nop",
						StrongID: "sha256:klm-id",
						Image: image.Image{
							Config: &runconfig.Config{
								Env:     []string{"d", "e", "f"},
								Volumes: map[string]struct{}{"vol2": struct{}{}},
							},
							Size: 2,
						},
					},
				},
			},
			"shared-layers": &distclient.Manifest{
				Layers: []distclient.Layer{
					{
						BlobSum:  "not-shared",
						StrongID: "sha256:not-shared",
					},
					{
						BlobSum:  "ghj-klm",
						StrongID: "sha256:ghj-id",
					},
				},
			},
		}

		blobs = map[digest.Digest]string{
			"abc-def": "abc-def-contents",
			"ghj-klm": "ghj-klm-contents",
			"klm-nop": "blah-blah",
		}

		fakeConn = new(fake_distclient.FakeConn)
		fakeConn.GetManifestStub = func(_ lager.Logger, tag string) (*distclient.Manifest, error) {
			return manifests[tag], nil
		}

		fakeConn.GetBlobReaderStub = func(_ lager.Logger, digest digest.Digest) (io.Reader, error) {
			return bytes.NewReader([]byte(blobs[digest])), nil
		}

		fakeDialer = new(fakes.FakeDialer)
		fakeDialer.DialStub = func(_ lager.Logger, host, repo, username, password string) (distclient.Conn, error) {
			return fakeConn, nil
		}

		fakeCake = new(fake_cake.FakeCake)
		fakeCake.GetStub = func(id layercake.ID) (*image.Image, error) {
			if _, ok := existingLayers[id.GraphID()]; ok {
				return &image.Image{Size: 33}, nil
			}

			return nil, errors.New("doesnt exist")
		}

		fakeVerifier = new(fakes.FakeVerifier)
		fakeVerifier.VerifyStub = func(r io.Reader, d digest.Digest) (io.ReadCloser, error) {
			return &verified{Reader: r}, nil
		}

		remote = repository_fetcher.NewRemote(defaultDockerRegistryHost, fakeCake, fakeDialer, fakeVerifier)
	})

	Context("when the URL has a host", func() {
		It("dials that host", func() {
			_, err := remote.Fetch(logger, parseURL("docker://some-host/some/repo#some-tag"), "", "", 1234)
			Expect(err).NotTo(HaveOccurred())

			_, host, _, _, _ := fakeDialer.DialArgsForCall(0)
			Expect(host).To(Equal("some-host"))
		})
	})

	Context("when the host is empty", func() {
		It("uses the default host", func() {
			_, err := remote.Fetch(logger, parseURL("docker:///some/repo#some-tag"), "", "", 1234)
			Expect(err).NotTo(HaveOccurred())

			_, host, _, _, _ := fakeDialer.DialArgsForCall(0)
			Expect(host).To(Equal(defaultDockerRegistryHost))
		})
	})

	Context("when the path contains a slash", func() {
		It("uses the path explicitly", func() {
			_, err := remote.Fetch(logger, parseURL("docker://some-host/some/repo#some-tag"), "", "", 1234)
			Expect(err).NotTo(HaveOccurred())

			_, _, repo, _, _ := fakeDialer.DialArgsForCall(0)
			Expect(repo).To(Equal("some/repo"))
		})
	})

	Context("when the path does not contain a slash", func() {
		Context("and the default registry is being used", func() {
			Context("and the default is DockerHub", func() {
				It("prepends the implied 'library/' to the path", func() {
					_, err := remote.Fetch(logger, parseURL("docker://registry-1.docker.io/somerepo#some-tag"), "", "", 1234)
					Expect(err).NotTo(HaveOccurred())

					_, _, repo, _, _ := fakeDialer.DialArgsForCall(0)
					Expect(repo).To(Equal("library/somerepo"))
				})
			})

			Context("and the default is a custom registry", func() {
				BeforeEach(func() {
					defaultDockerRegistryHost = "some-host"
				})

				It("does not prepend 'library/' to the path", func() {
					_, err := remote.Fetch(logger, parseURL("docker://some-host/somerepo#some-tag"), "", "", 1234)
					Expect(err).NotTo(HaveOccurred())

					_, _, repo, _, _ := fakeDialer.DialArgsForCall(0)
					Expect(repo).To(Equal("somerepo"))
				})
			})
		})

		Context("and a custom registry is being used", func() {
			It("does not prepend 'library/' to the path", func() {
				_, err := remote.Fetch(logger, parseURL("docker://some-host/somerepo#some-tag"), "", "", 1234)
				Expect(err).NotTo(HaveOccurred())

				_, _, repo, _, _ := fakeDialer.DialArgsForCall(0)
				Expect(repo).To(Equal("somerepo"))
			})
		})
	})

	Context("when the cake does not contain any of the layers", func() {
		JustBeforeEach(func() {
			_, err := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
			Expect(err).NotTo(HaveOccurred())
		})

		It("registers each of the layers in the graph", func() {
			Expect(fakeCake.RegisterCallCount()).To(Equal(3))
		})

		It("registers the layer contents under its Strong IDs", func() {
			image, reader := fakeCake.RegisterArgsForCall(0)
			Expect(image.ID).To(Equal("abc-id"))
			Expect(image.Parent).To(Equal("abc-parent-id"))

			b, err := ioutil.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal("abc-def-contents"))
		})

		It("registers the layer with the correct size", func() {
			image, _ := fakeCake.RegisterArgsForCall(0)
			Expect(image.Size).To(BeEquivalentTo(1))
		})
	})

	Context("when the graph already contains a layer", func() {
		JustBeforeEach(func() {
			existingLayers["ghj-id"] = true
		})

		It("avoids registering it again", func() {
			_, err := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCake.RegisterCallCount()).To(Equal(2))
		})
	})

	Context("when the url doesnot contain a fragment", func() {
		It("uses 'latest' as the tag", func() {
			_, err := remote.Fetch(logger, parseURL("docker:///foo"), "", "", 67)
			Expect(err).NotTo(HaveOccurred())

			_, tag := fakeConn.GetManifestArgsForCall(0)
			Expect(tag).To(Equal("latest"))
		})
	})

	It("returns an image with the ID of the top layer", func() {
		img, _ := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
		Expect(img.ImageID).To(Equal("klm-id"))
	})

	It("can fetch just the ID", func() {
		id, _ := remote.FetchID(logger, parseURL("docker:///foo#some-tag"))
		Expect(id).To(Equal(layercake.DockerImageID("klm-id")))
	})

	It("combines all the environment variable arrays together", func() {
		img, _ := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
		Expect(img.Env).To(ConsistOf([]string{"a", "b", "d", "e", "f"}))
	})

	It("combines all the volumes together", func() {
		img, _ := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
		Expect(img.Volumes).To(ConsistOf([]string{"vol1", "vol2"}))
	})

	It("should verify the image against its digest", func() {
		remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
		_, reader := fakeCake.RegisterArgsForCall(0)

		Expect(reader).To(BeAssignableToTypeOf(&verified{}))
		_, digest1 := fakeVerifier.VerifyArgsForCall(0)
		_, digest2 := fakeVerifier.VerifyArgsForCall(1)
		Expect(string(digest1)).To(Equal("abc-def"))
		Expect(string(digest2)).To(Equal("ghj-klm"))
	})

	It("should close the verified image reader after using it", func() {
		var registeredBlob *verified
		fakeCake.RegisterStub = func(img *image.Image, blob archive.ArchiveReader) error {
			Expect(blob).To(BeAssignableToTypeOf(&verified{}))
			Expect(blob.(*verified).closed).To(BeFalse())
			registeredBlob = blob.(*verified)
			return nil
		}

		remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
		Expect(registeredBlob.closed).To(BeTrue())
	})

	Context("when the layer does not match its digest", func() {
		JustBeforeEach(func() {
			fakeVerifier.VerifyReturns(nil, errors.New("boom"))
		})

		It("returns an error", func() {
			_, err := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
			Expect(err).To(MatchError("boom"))
		})

		It("does not register an image in the graph", func() {
			Expect(fakeCake.RegisterCallCount()).To(Equal(0))
		})
	})

	Describe("concurrently fetching", func() {
		It("serializes calls to cake.get and getblobreader", func() {
			for i := 1; i < 100; i++ {
				got := make(map[layercake.ID]bool)
				mu := new(sync.RWMutex)
				fakeCake.GetStub = func(id layercake.ID) (*image.Image, error) {
					mu.RLock()
					had := got[id]
					mu.RUnlock()

					if had {
						return nil, nil
					}

					return nil, errors.New("not found")
				}

				fakeCake.RegisterStub = func(img *image.Image, _ archive.ArchiveReader) error {
					mu.Lock()
					got[layercake.DockerImageID(img.ID)] = true
					mu.Unlock()

					return nil
				}

				wg := new(sync.WaitGroup)
				wg.Add(2)
				go func() {
					defer wg.Done()
					_, err := remote.Fetch(logger, parseURL("docker:///foo#some-tag"), "", "", 67)
					Expect(err).NotTo(HaveOccurred())
				}()

				go func() {
					defer wg.Done()
					_, err := remote.Fetch(logger, parseURL("docker:///foo#shared-layers"), "", "", 67)
					Expect(err).NotTo(HaveOccurred())
				}()

				wg.Wait()
				Expect(fakeConn.GetBlobReaderCallCount()).To(Equal(4 * i))
			}
		})
	})

	Context("when credentials are provided", func() {
		It("dials with the credentials", func() {
			_, err := remote.Fetch(logger, parseURL("docker:///banana#some-tag"), "username", "password", 3)
			Expect(err).NotTo(HaveOccurred())
			_, _, _, username, password := fakeDialer.DialArgsForCall(0)
			Expect(username).To(Equal("username"))
			Expect(password).To(Equal("password"))
		})

	})

	Context("when a disk quota is provided", func() {
		Context("and the image is smaller than the quota", func() {
			It("should succeed", func() {
				_, err := remote.Fetch(logger, parseURL("docker:///banana#some-tag"), "", "", 3)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("and the image is bigger than the quota", func() {
			It("should return an error", func() {
				_, err := remote.Fetch(logger, parseURL("docker:///banana#some-tag"), "", "", 2)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	It("returns the size of the image", func() {
		image, err := remote.Fetch(logger, parseURL("docker:///banana#some-tag"), "", "", 3)
		Expect(err).NotTo(HaveOccurred())
		Expect(image.Size).To(BeNumerically("==", 3))
	})
})

func parseURL(u string) *url.URL {
	r, err := url.Parse(u)
	Expect(err).NotTo(HaveOccurred())

	return r
}

type verified struct {
	io.Reader
	closed bool
}

func (v *verified) Close() error {
	v.closed = true
	return nil
}
