package retrier_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/garden-shed/pkg/retrier"
	"github.com/pivotal-golang/clock"
	"github.com/pivotal-golang/clock/fakeclock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Retrier", func() {
	var (
		ret             *retrier.Retrier
		clk             clock.Clock
		callbackCount   int
		callback        func() error
		timeout         time.Duration
		pollingInterval time.Duration
	)

	BeforeEach(func() {
		callbackCount = 0
		timeout = time.Millisecond * 240
		pollingInterval = time.Millisecond * 20
		callback = func() error {
			callbackCount++
			return nil
		}
		clk = clock.NewClock()
	})

	JustBeforeEach(func() {
		ret = &retrier.Retrier{
			Timeout:         timeout,
			PollingInterval: pollingInterval,
			Clock:           clk,
		}
	})

	Context("when the callback succeeds at first", func() {
		It("should call the callback only once", func() {
			Expect(ret.Retry(callback)).To(Succeed())
			Expect(callbackCount).To(Equal(1))
		})
	})

	Context("when the callback consistently fails", func() {
		BeforeEach(func() {
			callback = func() error {
				callbackCount++

				return errors.New("banana")
			}
		})

		It("should call the callback 12 times", func() {
			Expect(ret.Retry(callback)).NotTo(Succeed())
			Expect(callbackCount).To(Equal(12))
		})

		It("should return the error", func() {
			Expect(ret.Retry(callback)).To(MatchError("banana"))
		})
	})

	Context("when the callback succeeds after a while", func() {
		var (
			fakeClk *fakeclock.FakeClock
			called  chan bool
		)

		BeforeEach(func() {
			fakeClk = fakeclock.NewFakeClock(time.Now())
			clk = fakeClk
			called = make(chan bool, 12)

			callback = func() error {
				called <- true

				callbackCount++
				if callbackCount == 6 {
					return nil
				}

				return errors.New("acaboom")
			}
		})

		It("should honor the polling interval", func(done Done) {
			finished := make(chan struct{})
			go func(ret *retrier.Retrier) {
				defer GinkgoRecover()
				Expect(ret.Retry(callback)).To(Succeed())
				close(finished)
			}(ret)

			for i := 0; i < 5; i++ {
				Eventually(called).Should(Receive())
				fakeClk.WaitForWatcherAndIncrement(pollingInterval)
			}

			Eventually(finished).Should(BeClosed())
			close(done)
		}, 1.0)
	})
})
