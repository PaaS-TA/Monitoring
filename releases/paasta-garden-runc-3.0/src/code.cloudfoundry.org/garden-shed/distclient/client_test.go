package distclient_test

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/docker/docker/image"

	"code.cloudfoundry.org/garden-shed/distclient"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/docker/docker/runconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// busybox version to try to pull, should be a tag so it doesn't change
const busyBoxVersion = "1.24.0"

// expected busybox layer digests (these should never change since the tag above is locked down)
var busyBoxLayers = []distclient.Layer{
	{
		BlobSum:  "sha256:1b373b69cd34f11679c6059262b6fed80eeb7b38ca3e257e3f689c1aaba6df54",
		StrongID: "sha256:ab2b8a86ca6c4be761a7128150fa6220735bdb277555082912ca4b26fd3ae264",
		Image: image.Image{
			Config: &runconfig.Config{
				Env: []string{"a", "b"},
			},
		},
	},
	{
		BlobSum:        "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4",
		StrongID:       "sha256:2c5ac3f849df8627fcf2822727f87c57f38b7129d3604fbc11d861fe856ff093",
		ParentStrongID: "sha256:ab2b8a86ca6c4be761a7128150fa6220735bdb277555082912ca4b26fd3ae264",
		Image: image.Image{
			Config: &runconfig.Config{},
		},
	},
}

var busyBoxLayerContents = [][]string{
	[]string{"bin", "dev", "etc", "home", "root", "tmp", "usr", "var"},
	[]string{},
}

var _ = Describe("distclient", func() {
	var (
		logger lager.Logger
		conn   distclient.Conn
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		d := distclient.NewDialer([]string{})

		var err error
		conn, err = d.Dial(logger, "registry-1.docker.io", "library/busybox", "", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("can pull a manifest from dockerhub", func() {
		layer, err := conn.GetManifest(logger, busyBoxVersion)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Layers[0].BlobSum).To(Equal(busyBoxLayers[0].BlobSum))
		Expect(layer.Layers[1].BlobSum).To(Equal(busyBoxLayers[1].BlobSum))

		Expect(layer.Layers[0].StrongID).To(Equal(busyBoxLayers[0].StrongID))
		Expect(layer.Layers[1].StrongID).To(Equal(busyBoxLayers[1].StrongID))

		Expect(layer.Layers[0].ParentStrongID).To(Equal(busyBoxLayers[0].ParentStrongID))
		Expect(layer.Layers[1].ParentStrongID).To(Equal(busyBoxLayers[1].ParentStrongID))

		Expect(layer.Layers[0].Image.ContainerConfig.Env).To(Equal(busyBoxLayers[0].Image.ContainerConfig.Env))
		Expect(layer.Layers[1].Image.ContainerConfig.Env).To(Equal(busyBoxLayers[1].Image.ContainerConfig.Env))
	})

	It("returns bottom layer to top layer (reverse of docker api, order they should be applied to the graph)", func() {
		layer, err := conn.GetManifest(logger, busyBoxVersion)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.Layers[0].ParentStrongID).To(BeEquivalentTo(""))
	})

	It("can get a layer blob from dockerhub", func() {
		for i, layer := range busyBoxLayers {
			tmp := tmpDir()
			defer os.RemoveAll(tmp)

			r, err := conn.GetBlobReader(logger, layer.BlobSum)
			Expect(err).NotTo(HaveOccurred())

			cmd := exec.Command("tar", "zxf", "-", "-C", tmp)
			cmd.Stdin = r

			tarSession, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(tarSession, "30s").Should(gexec.Exit(0))
			Expect(fileNames(tmp)).To(ConsistOf(busyBoxLayerContents[i]))
		}
	})
})

func tmpDir() string {
	tmp, err := ioutil.TempDir("", "")
	Expect(err).NotTo(HaveOccurred())
	return tmp
}

func fileNames(path string) (names []string) {
	dir, err := ioutil.ReadDir(path)
	Expect(err).NotTo(HaveOccurred())

	for _, d := range dir {
		names = append(names, d.Name())
	}

	return
}
