package garden_integration_tests_test

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	gardenHost            string
	gardenPort            string
	gardenDebugPort       string
	gardenClient          garden.Client
	container             garden.Container
	containerCreateErr    error
	assertContainerCreate bool

	handle              string
	imageRef            garden.ImageRef
	networkSpec         string
	privilegedContainer bool
	properties          garden.Properties
	limits              garden.Limits
	env                 []string
	ginkgoIO            garden.ProcessIO = garden.ProcessIO{
		Stdout: GinkgoWriter,
		Stderr: GinkgoWriter,
	}
)

func TestGardenIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(5 * time.Second)

	BeforeEach(func() {
		assertContainerCreate = true
		handle = ""
		imageRef = garden.ImageRef{}
		networkSpec = ""
		privilegedContainer = false
		properties = garden.Properties{}
		limits = garden.Limits{}
		env = []string{}
		gardenHost = os.Getenv("GARDEN_ADDRESS")
		if gardenHost == "" {
			gardenHost = "10.244.16.6"
		}
		gardenPort = os.Getenv("GARDEN_PORT")
		if gardenPort == "" {
			gardenPort = "7777"
		}
		gardenDebugPort = os.Getenv("GARDEN_DEBUG_PORT")
		if gardenDebugPort == "" {
			gardenDebugPort = "17013"
		}
		gardenClient = client.New(connection.New("tcp", fmt.Sprintf("%s:%s", gardenHost, gardenPort)))
	})

	JustBeforeEach(func() {
		container, containerCreateErr = gardenClient.Create(garden.ContainerSpec{
			Handle:     handle,
			Image:      imageRef,
			Privileged: privilegedContainer,
			Properties: properties,
			Env:        env,
			Limits:     limits,
			Network:    networkSpec,
		})

		if container != nil {
			fmt.Fprintf(GinkgoWriter, "Container handle: %s\n", container.Handle())
		}

		if assertContainerCreate {
			Expect(containerCreateErr).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		if container != nil {
			// ignoring the error since it can return unknown handle error
			theContainer, _ := gardenClient.Lookup(container.Handle())

			if theContainer != nil {
				Expect(gardenClient.Destroy(container.Handle())).To(Succeed())
			}
		}
	})

	RunSpecs(t, "GardenIntegrationTests Suite")
}

func getContainerHandles() []string {
	containers, err := gardenClient.Containers(nil)
	Expect(err).ToNot(HaveOccurred())

	handles := make([]string, len(containers))
	for i, c := range containers {
		handles[i] = c.Handle()
	}

	return handles
}

func createUser(container garden.Container, username string) {
	if container == nil {
		return
	}

	process, err := container.Run(garden.ProcessSpec{
		User: "root",
		Path: "sh",
		Args: []string{"-c", fmt.Sprintf("id -u %s || adduser -D %s", username, username)},
	}, garden.ProcessIO{
		Stdout: GinkgoWriter,
		Stderr: GinkgoWriter,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(process.Wait()).To(Equal(0))
}

func getKernelVersion() (int, int) {
	container, err := gardenClient.Create(garden.ContainerSpec{})
	Expect(err).NotTo(HaveOccurred())
	defer gardenClient.Destroy(container.Handle())

	var outBytes bytes.Buffer
	process, err := container.Run(garden.ProcessSpec{
		User: "root",
		Path: "uname",
		Args: []string{"-r"},
	}, garden.ProcessIO{
		Stdout: &outBytes,
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(process.Wait()).To(Equal(0))

	vSplit := strings.Split(outBytes.String(), ".")
	major, err := strconv.Atoi(vSplit[0])
	Expect(err).NotTo(HaveOccurred())
	minor, err := strconv.Atoi(vSplit[1])
	Expect(err).NotTo(HaveOccurred())

	return major, minor
}
