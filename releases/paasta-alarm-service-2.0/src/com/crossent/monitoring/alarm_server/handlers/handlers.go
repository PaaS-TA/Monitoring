package handlers

import (
	"os"
	"github.com/tedsuo/ifrit"
	"com/crossent/monitoring/alarm_server/services"
)

type alarm_server struct{
	config 	*services.AlarmConfig
}



func NewHandler(alarm_config *services.AlarmConfig) ifrit.Runner {
	return &alarm_server{
		config: alarm_config,
	}
}

func (n *alarm_server) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	close(ready)

	//===============================================================
	// Call Service
	caller := services.NewAlarmService(n.config)
	err := caller.CheckSystemThreshold()
	//===============================================================
	return err
}