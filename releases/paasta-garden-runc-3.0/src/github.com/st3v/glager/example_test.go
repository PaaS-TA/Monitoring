package glager_test

import (
	"errors"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	. "github.com/st3v/glager"
)

func myFunc(logger lager.Logger) {
	logger.Info("myFunc", lager.Data(map[string]interface{}{
		"event": "start",
	}))

	logger.Debug("myFunc", lager.Data(map[string]interface{}{
		"some": "stuff",
		"more": "stuff",
	}))

	logger.Error("myFunc", errors.New("some-err"), lager.Data(map[string]interface{}{
		"details": "stuff",
	}))

	logger.Info("myFunc", lager.Data(map[string]interface{}{
		"event": "done",
	}))
}

var _ = Describe("HaveLogged", func() {
	It("can be used like this", func() {
		logger := NewLogger("test")

		myFunc(logger)

		Expect(logger).To(HaveLogged(
			Info(
				Message("test.myFunc"),
				Data("event", "start"),
			),
			Info(
				Message("test.myFunc"),
				Data("event", "done"),
			),
		))
	})
})

var _ = Describe("ContainSequence", func() {
	It("can be used like this", func() {
		log := gbytes.NewBuffer()
		logger := lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(log, lager.DEBUG))

		myFunc(logger)

		Expect(log).To(ContainSequence(
			Info(Data("event", "start")),
			Error(AnyErr),
		))
	})
})
