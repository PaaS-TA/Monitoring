package glager_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/types"

	. "github.com/st3v/glager"
)

func TestGlager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Glager Test Suite")
}

var _ = Describe(".HaveLogged", func() {
	var (
		logger            lager.Logger
		expectedSource    = "some-source"
		action            = "some-action"
		expectedAction    = fmt.Sprintf("%s.%s", expectedSource, action)
		expectedDataKey   = "some-key"
		expectedDataValue = "some-value"
	)

	BeforeEach(func() {
		logger = NewLogger(expectedSource)
		logger.Info(action, lager.Data{expectedDataKey: expectedDataValue})
	})

	It("matches an entry", func() {
		Expect(logger).To(ContainSequence(
			Info(
				Action(expectedAction),
				Data(expectedDataKey, expectedDataValue),
			),
		))
	})
})

var _ = Describe(".ContainSequence", func() {
	var (
		logger         lager.Logger
		expectedSource = "some-source"
	)

	BeforeEach(func() {
		logger = NewLogger(expectedSource)
	})

	Context("when actual contains an entry", func() {
		var (
			action            = "some-action"
			expectedAction    = fmt.Sprintf("%s.%s", expectedSource, action)
			expectedDataKey   = "some-key"
			expectedDataValue = "some-value"
		)

		Context("that is an info", func() {
			BeforeEach(func() {
				logger.Info(action, lager.Data{expectedDataKey: expectedDataValue})
			})

			It("matches an empty info entry", func() {
				Expect(logger).To(ContainSequence(
					Info(),
				))
			})

			It("matches an info entry with a source only", func() {
				Expect(logger).To(ContainSequence(
					Info(
						Source(expectedSource),
					),
				))
			})

			It("matches an info entry with a message only", func() {
				Expect(logger).To(ContainSequence(
					Info(
						Message(expectedAction),
					),
				))
			})

			It("matches an info entry with an action only", func() {
				Expect(logger).To(ContainSequence(
					Info(
						Action(expectedAction),
					),
				))
			})

			It("matches an info entry with data only", func() {
				Expect(logger).To(ContainSequence(
					Info(
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("matches the correct info entry", func() {
				Expect(logger).To(ContainSequence(
					Info(
						Source(expectedSource),
						Message(expectedAction),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match an info entry with an incorrect source", func() {
				Expect(logger).ToNot(ContainSequence(
					Info(
						Source("invalid"),
						Message(expectedAction),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match an info entry with an incorrect message", func() {
				Expect(logger).ToNot(ContainSequence(
					Info(
						Source(expectedSource),
						Message("invalid"),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match an info entry with incorrect data", func() {
				Expect(logger).ToNot(ContainSequence(
					Info(
						Source(expectedSource),
						Message(expectedAction),
						Data(expectedDataKey, expectedDataValue, "non-existing-key", "non-existing-value"),
					),
				))
			})

			It("does not match a debug entry", func() {
				Expect(logger).ToNot(ContainSequence(Debug()))
			})

			It("does not match an error entry", func() {
				Expect(logger).ToNot(ContainSequence(Error(AnyErr)))
			})

			It("does not match a fatal entry", func() {
				Expect(logger).ToNot(ContainSequence(Fatal(AnyErr)))
			})

			Context("with non-string data values", func() {
				type foo struct {
					Foo string `json:"foo"`
				}

				var (
					obj foo
					arr []string
				)

				BeforeEach(func() {
					obj = foo{"bar"}

					arr = []string{"a", "b", "c"}

					logger.Info("non-string-data", lager.Data{
						"int":    17,
						"float":  1.23,
						"bool":   true,
						"array":  arr,
						"object": obj,
						"null":   nil,
					})
				})

				It("matches a correct int value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("int", 17)),
					))
				})

				It("does not match an incorrect int value", func() {
					Expect(logger).ToNot(ContainSequence(
						Info(Data("int", 99)),
					))
				})

				It("matches a correct float value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("float", 1.23)),
					))
				})

				It("does not match an incorrect float value", func() {
					Expect(logger).ToNot(ContainSequence(
						Info(Data("float", 1.2)),
					))
				})

				It("matches a correct bool value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("bool", true)),
					))
				})

				It("does not match an incorrect bool value", func() {
					Expect(logger).ToNot(ContainSequence(
						Info(Data("bool", false)),
					))
				})

				It("matches a correct null value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("null", nil)),
					))
				})

				It("matches a correct array value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("array", arr)),
					))
				})

				It("does not match an incorrect array value", func() {
					Expect(logger).ToNot(ContainSequence(
						Info(Data("array", []string{"a", "b"})),
					))

					Expect(logger).ToNot(ContainSequence(
						Info(Data("array", []int{1})),
					))
				})

				It("matches a correct object value", func() {
					Expect(logger).To(ContainSequence(
						Info(Data("object", obj)),
					))
				})

				It("does not match an incorrect object value", func() {
					Expect(logger).ToNot(ContainSequence(
						Info(Data("object", foo{"boooh"})),
					))

					Expect(logger).ToNot(ContainSequence(
						Info(Data("object", errors.New("something"))),
					))
				})
			})
		})

		Context("that is an error", func() {
			var expectedErr = errors.New("some-error")

			BeforeEach(func() {
				logger.Error(action, expectedErr, lager.Data{expectedDataKey: expectedDataValue})
			})

			It("does match the correct error without additional fields", func() {
				Expect(logger).To(ContainSequence(
					Error(
						expectedErr,
					),
				))
			})

			It("does match AnyErr", func() {
				Expect(logger).To(ContainSequence(
					Error(
						AnyErr,
					),
				))
			})

			It("does match nil err", func() {
				Expect(logger).To(ContainSequence(
					Error(
						nil,
					),
				))
			})

			It("does match the correct error with correct additional fields", func() {
				Expect(logger).To(ContainSequence(
					Error(
						expectedErr,
						Source(expectedSource),
						Action(expectedAction),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match an incorrect error", func() {
				Expect(logger).ToNot(ContainSequence(Error(errors.New("some-other-error"))))
			})

			It("does not match the correct error with incorrect source", func() {
				Expect(logger).ToNot(ContainSequence(
					Error(
						expectedErr,
						Source("incorrect"),
					),
				))
			})

			It("does not match the correct error with incorrect message", func() {
				Expect(logger).ToNot(ContainSequence(
					Error(
						expectedErr,
						Message("incorrect"),
					),
				))
			})

			It("does not match the correct error with incorrect data", func() {
				Expect(logger).ToNot(ContainSequence(
					Error(
						expectedErr,
						Data("non-exiting-key", "non-existing-value"),
					),
				))
			})

			It("does not match an info entry", func() {
				Expect(logger).ToNot(ContainSequence(Info()))
			})

			It("does not match a debug entry", func() {
				Expect(logger).ToNot(ContainSequence(Debug()))
			})

			It("does not match a fatal entry", func() {
				Expect(logger).ToNot(ContainSequence(Fatal(AnyErr)))
			})
		})

		Context("that is a debug entry", func() {
			BeforeEach(func() {
				logger.Debug(action, lager.Data{expectedDataKey: expectedDataValue})
			})

			It("does match an empty debug entry", func() {
				Expect(logger).To(ContainSequence(Debug()))
			})

			It("does match the correct debug entry", func() {
				Expect(logger).To(ContainSequence(
					Debug(
						Source(expectedSource),
						Message(expectedAction),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match a debug entry with an incorrect source", func() {
				Expect(logger).ToNot(ContainSequence(
					Debug(
						Source("incorrect"),
					),
				))
			})

			It("does not match a debug entry with an incorrect message", func() {
				Expect(logger).ToNot(ContainSequence(
					Debug(
						Message("incorrect"),
					),
				))
			})

			It("does not match a debug entry with a incorrect data", func() {
				Expect(logger).ToNot(ContainSequence(
					Debug(
						Data("non-existing-key"),
					),
				))
			})

			It("does not match an info entry", func() {
				Expect(logger).ToNot(ContainSequence(Info()))
			})

			It("does not match an error entry", func() {
				Expect(logger).ToNot(ContainSequence(Error(AnyErr)))
			})

			It("does not match a fatal entry", func() {
				Expect(logger).ToNot(ContainSequence(Fatal(AnyErr)))
			})
		})

		Context("that is a fatal error", func() {
			var expectedErr = errors.New("some-error")

			BeforeEach(func() {
				Expect(func() {
					logger.Fatal(action, expectedErr, lager.Data{expectedDataKey: expectedDataValue})
				}).To(Panic())
			})

			It("does match fatal entry with AnyErr", func() {
				Expect(logger).To(ContainSequence(Fatal(AnyErr)))
			})

			It("does match fatal entry with nil err", func() {
				Expect(logger).To(ContainSequence(Fatal(nil)))
			})

			It("does match a fatal entry with correct error", func() {
				Expect(logger).To(ContainSequence(
					Fatal(
						expectedErr,
					),
				))
			})

			It("does match a fatal entry with correct error and additional fields", func() {
				Expect(logger).To(ContainSequence(
					Fatal(
						expectedErr,
						Source(expectedSource),
						Message(expectedAction),
						Data(expectedDataKey, expectedDataValue),
					),
				))
			})

			It("does not match a fatal entry with an incorrect error", func() {
				Expect(logger).ToNot(ContainSequence(
					Fatal(
						errors.New("some-other-error"),
					),
				))
			})

			It("does not match a fatal entry with an incorrect source", func() {
				Expect(logger).ToNot(ContainSequence(
					Fatal(
						expectedErr,
						Source("incorrect"),
					),
				))
			})

			It("does not match a fatal entry with an incorrect action", func() {
				Expect(logger).ToNot(ContainSequence(
					Fatal(
						expectedErr,
						Action("incorrect"),
					),
				))
			})

			It("does not match a fatal entry with incorrect data", func() {
				Expect(logger).ToNot(ContainSequence(
					Fatal(
						expectedErr,
						Data("incorrect"),
					),
				))
			})

			It("does not match an info entry", func() {
				Expect(logger).ToNot(ContainSequence(Info()))
			})

			It("does not match a debug entry", func() {
				Expect(logger).ToNot(ContainSequence(Debug()))
			})

			It("does not match an error entry", func() {
				Expect(logger).ToNot(ContainSequence(Error(AnyErr)))
			})
		})
	})

	Context("when actual contains multiple entries", func() {
		var expectedError = errors.New("some-error")

		BeforeEach(func() {
			logger.Info("action", lager.Data{"event": "starting", "task": "my-task"})
			logger.Debug("action", lager.Data{"event": "debugging", "task": "my-task"})
			logger.Error("action", expectedError, lager.Data{"event": "failed", "task": "my-task"})
		})

		It("does match a correct sequence", func() {
			Expect(logger).To(ContainSequence(
				Info(
					Data("event", "starting", "task", "my-task"),
				),
				Debug(
					Data("event", "debugging", "task", "my-task"),
				),
				Error(
					expectedError,
					Data("event", "failed", "task", "my-task"),
				),
			))
		})

		It("does match a correct subsequence with missing elements in the beginning", func() {
			Expect(logger).To(ContainSequence(
				Debug(
					Data("event", "debugging", "task", "my-task"),
				),
				Error(
					expectedError,
					Data("event", "failed", "task", "my-task"),
				),
			))
		})

		It("does match a correct subsequence with missing elements in the end", func() {
			Expect(logger).To(ContainSequence(
				Info(
					Data("event", "starting", "task", "my-task"),
				),
				Debug(
					Data("event", "debugging", "task", "my-task"),
				),
			))
		})

		It("does match a correct but non-continious subsequence", func() {
			Expect(logger).To(ContainSequence(
				Info(
					Data("event", "starting", "task", "my-task"),
				),
				Error(
					expectedError,
					Data("event", "failed", "task", "my-task"),
				),
			))
		})

		It("does not match an incorrect sequence", func() {
			Expect(logger).ToNot(ContainSequence(
				Info(
					Data("event", "starting", "task", "my-task"),
				),
				Info(
					Data("event", "starting", "task", "my-task"),
				),
			))
		})

		It("does not match an out-of-order sequence", func() {
			Expect(logger).ToNot(ContainSequence(
				Debug(
					Data("event", "debugging", "task", "my-task"),
				),
				Error(
					expectedError,
					Data("event", "failed", "task", "my-task"),
				),
				Info(
					Data("event", "starting", "task", "my-task"),
				),
			))
		})
	})

	Describe("logMatcher", func() {
		var (
			buffer  *gbytes.Buffer
			logger  lager.Logger
			matcher types.GomegaMatcher
		)

		BeforeEach(func() {
			matcher = ContainSequence(Info())
			buffer = gbytes.NewBuffer()
			logger = lager.NewLogger("logger")
			logger.RegisterSink(lager.NewWriterSink(buffer, lager.DEBUG))
			logger.Debug("some-debug")
		})

		Describe("Match", func() {
			var (
				actual  interface{}
				success bool
				err     error
			)

			JustBeforeEach(func() {
				logger.Info("some-info")
				success, err = matcher.Match(actual)
			})

			Context("when actual is an invalid type", func() {
				BeforeEach(func() {
					actual = "invalid"
				})

				It("returns failure", func() {
					Expect(success).To(BeFalse())
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("ContainSequence must be passed"))
				})
			})

			Context("when actual is a BufferProvider", func() {
				var sink *lagertest.TestSink

				BeforeEach(func() {
					sink = lagertest.NewTestSink()
					logger.RegisterSink(sink)
					actual = sink
				})

				It("returns success", func() {
					Expect(success).To(BeTrue())
				})

				It("does not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("does match on subsequent calls", func() {
					Expect(actual).To(matcher)
				})
			})

			Context("when actual is a ContentsProvider", func() {
				BeforeEach(func() {
					actual = buffer
				})

				It("returns success", func() {
					Expect(success).To(BeTrue())
				})

				It("does not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("does match on subsequent calls", func() {
					Expect(actual).To(matcher)
				})
			})

			Context("when actual is an io.Reader", func() {
				BeforeEach(func() {
					actual = bufio.NewReader(buffer)
				})

				It("returns success", func() {
					Expect(success).To(BeTrue())
				})

				It("does not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not match on subsequent calls", func() {
					Expect(actual).ToNot(matcher)
				})
			})

			Context("when actual contains invalid entries", func() {
				BeforeEach(func() {
					actual = strings.NewReader("invalid")
				})

				It("returns failure", func() {
					Expect(success).To(BeFalse())
				})

				It("returns a json.SyntaxError", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(BeAssignableToTypeOf(&json.SyntaxError{}))
				})
			})

			Context("when expected contains non-encodable data values", func() {
				BeforeEach(func() {
					actual = buffer
					logger.Info("foo", lager.Data{"foo": "bar"})
					matcher = ContainSequence(Info(Data("foo", func() {})))
				})

				It("returns a json.UnsupportedType error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(BeAssignableToTypeOf(&json.UnsupportedTypeError{}))
				})
			})
		})

		Describe("FailureMessage", func() {
			It("returns the right message", func() {
				matcher.Match(buffer)
				Expect(matcher.FailureMessage(buffer)).To(ContainSubstring(
					"to contain log sequence",
				))
			})
		})

		Describe("NegatedFailureMessage", func() {
			It("returns the right message", func() {
				matcher.Match(buffer)
				Expect(matcher.NegatedFailureMessage(buffer)).To(ContainSubstring(
					"not to contain log sequence",
				))
			})
		})
	})

	Describe(".Data", func() {
		Context("when a non-string key is passed", func() {
			It("panics", func() {
				Expect(func() {
					Info(Data([]string{"foo"}, "bar"))
				}).To(Panic())
			})
		})
	})
})
