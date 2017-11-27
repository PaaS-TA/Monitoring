package repository_fetcher_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/ioutil"

	"code.cloudfoundry.org/garden-shed/repository_fetcher"
	"github.com/docker/distribution/digest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const someShaThatDoesntMatch = "sha256:df3ae2b606ca0ab01a4bc6ec2b7450a547106b47eca44a242153d3bb3fc254b9"

var shaThatDoesMatch = digest.Digest(fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("matches"))))

var _ = Describe("Verifying a digest", func() {
	Context("when the digest matches", func() {
		It("does not return an error", func() {
			_, err := repository_fetcher.Verify(bytes.NewReader([]byte("matches")), shaThatDoesMatch)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows reading the original data", func() {
			r, err := repository_fetcher.Verify(bytes.NewReader([]byte("matches")), shaThatDoesMatch)
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.ReadAll(r)).To(Equal([]byte("matches")))
		})
	})

	Context("when the digest does not match", func() {
		It("returns an error", func() {
			_, err := repository_fetcher.Verify(bytes.NewReader([]byte("does-not-match")), someShaThatDoesntMatch)
			Expect(err).To(MatchError("digest verification failed"))
		})
	})
})
