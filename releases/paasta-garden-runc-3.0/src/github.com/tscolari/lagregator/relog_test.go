package lagregator_test

import (
	"errors"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/st3v/glager"
	"github.com/tscolari/lagregator"
)

var _ = Describe("Relog", func() {
	var (
		destLogger lager.Logger
		srcLogger  lager.Logger
		srcBuffer  *gbytes.Buffer
	)

	BeforeEach(func() {
		destLogger = glager.NewLogger("destination")
		srcLogger = glager.NewLogger("source")
		srcBuffer = gbytes.NewBuffer()
		srcSink := lager.NewWriterSink(srcBuffer, lager.DEBUG)
		srcLogger.RegisterSink(srcSink)
	})

	Describe("RelogBytes", func() {
		It("relogs from another lagger output", func() {
			srcLogger.Debug("first-debug", lager.Data{"attr1": "value1"})
			srcLogger.Info("first-info", lager.Data{"attr1": "value1"})
			srcLogger.Error("first-error", errors.New("failed!"), lager.Data{"attr1": "value1"})
			srcLogger.Debug("second-debug", lager.Data{"attr2": "value2"})

			lagregator.RelogBytes(destLogger, srcBuffer.Contents())

			Expect(destLogger).To(glager.HaveLogged(
				glager.Debug(
					glager.Source("destination"),
					glager.Message("destination.source.first-debug"),
					glager.Data("attr1", "value1"),
				),
				glager.Info(
					glager.Source("destination"),
					glager.Message("destination.source.first-info"),
					glager.Data("attr1", "value1"),
				),
				glager.Error(
					errors.New("failed!"),
					glager.Source("destination"),
					glager.Message("destination.source.first-error"),
					glager.Data("attr1", "value1"),
				),
				glager.Debug(
					glager.Source("destination"),
					glager.Message("destination.source.second-debug"),
					glager.Data("attr2", "value2"),
				),
			))
		})
	})

	Describe("RelogStream", func() {
		It("relogs from another lager stream", func() {
			srcLogger.Debug("first-debug", lager.Data{"attr1": "value1"})
			srcLogger.Info("first-info", lager.Data{"attr1": "value1"})
			srcLogger.Error("first-error", errors.New("failed!"), lager.Data{"attr1": "value1"})
			srcLogger.Debug("second-debug", lager.Data{"attr2": "value2"})

			lagregator.RelogStream(destLogger, srcBuffer)

			Expect(destLogger).To(glager.HaveLogged(
				glager.Debug(
					glager.Source("destination"),
					glager.Message("destination.source.first-debug"),
					glager.Data("attr1", "value1"),
				),
				glager.Info(
					glager.Source("destination"),
					glager.Message("destination.source.first-info"),
					glager.Data("attr1", "value1"),
				),
				glager.Error(
					errors.New("failed!"),
					glager.Source("destination"),
					glager.Message("destination.source.first-error"),
					glager.Data("attr1", "value1"),
				),
				glager.Debug(
					glager.Source("destination"),
					glager.Message("destination.source.second-debug"),
					glager.Data("attr2", "value2"),
				),
			))
		})
	})
})
