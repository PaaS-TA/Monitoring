package services_test

import(
	"time"
	"com/crossent/monitoring_agent/services"
	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
var _ = Describe("RuntimeSystemStats", func() {
	var (
		fakeSender *services.MetricSender
		fakeLogger lager.Logger
		fakeInflux *services.InfluxConfig
		origin string
		cellIp string
		runDone chan struct{}
		stopChan chan bool
	)

	BeforeEach(func() {
		origin = "localhost"
		cellIp = "127.0.0.l"
		fakeInflux = new(services.InfluxConfig)
		fakeInflux.InfluxUrl = "127.0.0.1:8059"
		fakeInflux.InfluxDatabase = "bosh_metric_db"
		fakeInflux.Measurement = "bosh_metrics"
		fakeSender = services.NewMetricSender(fakeLogger, fakeInflux, origin, cellIp, 5*time.Second)
		stopChan = make(chan bool)
		runDone = make(chan struct{})
	})

	AfterEach(func() {
		close(stopChan)
		Eventually(runDone).Should(BeClosed())
	})

	var perform = func() {
		go func() {
			fakeSender.SendMetricsToInfluxDb(stopChan)
			close(runDone)
		}()
	}

	It("periodically collect metrics on Bosh Service Environment", func() {
		perform()

		Eventually(func() bool {
			return fakeSender.Success
		}, 5*time.Second, 1*time.Second).Should(BeTrue())
		//Eventually(fakeSender.Success).Should(BeTrue())
	})
})