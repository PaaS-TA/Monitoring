package handler

import (
	"os"

	"com/crossent/monitoring_agent/services"
	"code.cloudfoundry.org/lager"

	"github.com/tedsuo/ifrit"
	"time"
)

const (
	statsInterval        = 30 * time.Second
)


type metrics_sender_server struct {
	logger		lager.Logger
	influxCon 	*services.InfluxConfig
	origin 		string
	cellIp 		string
}

func New(logger lager.Logger, influxCon *services.InfluxConfig, origin, cellIp string) ifrit.Runner {
	return &metrics_sender_server{
		logger: logger,
		influxCon: influxCon,
		origin: origin,
		cellIp: cellIp,
	}
}


func (n *metrics_sender_server) Run(signals <-chan os.Signal, ready chan<- struct{}) error {

	//===============================================================
	// Call Service
	metrics_sender := services.NewMetricSender(n.logger, n.influxCon, n.origin, n.cellIp, statsInterval)
	err := metrics_sender.SendMetricsToInfluxDb(nil)
	//===============================================================
	return err
}
