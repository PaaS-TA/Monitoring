package lagregator_test

import (
	"bytes"
	"errors"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/st3v/glager"
	"github.com/tscolari/lagregator"
)

var _ = Describe("Relogger", func() {
	var (
		destLogger lager.Logger
		srcLogger  lager.Logger
	)

	BeforeEach(func() {
		destLogger = glager.NewLogger("destination")
		srcLogger = glager.NewLogger("source")
	})

	It("returns an io.Writer that relogs to destination", func() {
		relogger := lagregator.NewRelogger(destLogger)
		srcLogger.RegisterSink(lager.NewWriterSink(relogger, lager.DEBUG))

		srcLogger.Debug("first-debug", lager.Data{"attr1": "value1"})
		srcLogger.Info("first-info", lager.Data{"attr1": "value1"})
		srcLogger.Error("first-error", errors.New("failed!"), lager.Data{"attr1": "value1"})
		srcLogger.Debug("second-debug", lager.Data{"attr2": "value2"})

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

	Context("when the relogger receives not only valid lager output", func() {
		It("ignores anything that doesn't look like lager logs", func() {
			destBuffer := gbytes.NewBuffer()
			destLogger.RegisterSink(lager.NewWriterSink(destBuffer, lager.DEBUG))

			relogger := lagregator.NewRelogger(destLogger)
			srcSink := lager.NewWriterSink(relogger, lager.DEBUG)
			srcLogger.RegisterSink(srcSink)

			srcLogger.Debug("first-debug", lager.Data{"attr1": "value1"})
			srcLogger.Info("first-info", lager.Data{"attr1": "value1"})

			_, err := relogger.Write([]byte("not a cool thing to have in the logs"))
			Expect(err).ToNot(HaveOccurred())

			srcLogger.Error("first-error", errors.New("failed!"), lager.Data{"attr1": "value1"})
			_, err = relogger.Write([]byte("\n\n {} \n\n wrong stuff"))
			Expect(err).ToNot(HaveOccurred())
			srcLogger.Debug("second-debug", lager.Data{"attr2": "value2"})

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

			Expect(destBuffer).ToNot(gbytes.Say("not a cool thing to have in the logs"))
			Expect(destBuffer).ToNot(gbytes.Say("wront stuff"))
		})
	})

	Describe("Performance of lager.Logger", func() {
		Context("Using buffer", func() {
			var logger lager.Logger
			BeforeEach(func() {
				logger = lager.NewLogger("test")
				logger.RegisterSink(lager.NewWriterSink(bytes.NewBuffer([]byte{}), lager.DEBUG))
			})

			Measure("lager.Logger using buffer", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					logger.Info("logging message", lager.Data{"key": "value"})
				})

				b.RecordValue("Time took", runtime.Seconds())
			}, 10000)
		})

		Context("Using relogger", func() {
			BeforeEach(func() {
				relogger := lagregator.NewRelogger(destLogger)
				srcLogger.RegisterSink(lager.NewWriterSink(relogger, lager.DEBUG))
			})

			Measure("lager.Logger using buffer", func(b Benchmarker) {
				runtime := b.Time("runtime", func() {
					srcLogger.Info("logging message", lager.Data{"key": "value"})
				})

				b.RecordValue("Time took", runtime.Seconds())
			}, 10000)
		})
	})
})
