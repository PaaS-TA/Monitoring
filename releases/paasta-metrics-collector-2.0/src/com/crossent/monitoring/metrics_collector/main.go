package main

import (
	"os"
	"strings"
	"time"
	"flag"
	"io/ioutil"
	"strconv"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-incubator/runtime-schema/cc_messages/flags"
//	"github.com/cloudfoundry/dropsonde"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/ifrit/grouper"

	"com/crossent/monitoring/metrics_collector/services"
	"com/crossent/monitoring/metrics_collector/handler"
	"com/crossent/monitoring/metrics_collector/util"
)
var clientId = flag.String(
	"clientId",
	"",
	"UAA client id for doppler service ",
)

var clientPass = flag.String(
	"clientPass",
	"",
	"UAA client password for doppler service ",
)
var uaaUrl = flag.String(
	"uaaUrl",
	"",
	"Address of UAA ",
)

var influxUrl = flag.String(
	"influxUrl",
	"",
	"Address of Influx Time Series Database ",
)

var influxDatabase = flag.String(
	"influxDatabase",
	"",
	"Influx Database name",
)

var cfMeasurement = flag.String(
	"cfMeasurement",
	"",
	"Influx CF Metrics Measurement name",
)

var cfProcessMeasurement = flag.String(
	"cfProcessMeasurement",
	"",
	"Influx CF Process Measurement name",
)

var dopplerUrl = flag.String(
	"dopplerUrl",
	"",
	"doppler url. In case of multi doppler url, Do separate info by comma. example wss://doppler.bosh-lite.com,wss://doppler2.bosh-lite.com  ",
)

var pidFile = flag.String(
	"pidFile",
	"",
	"File for Current Process ID",
)

type Config map[string]string

/*const (
	dropsondeOrigin      = "metrics_collector"
	dropsondeDestination = "localhost:3457"
)*/

func main() {
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)

	lifecycles := flags.LifecycleMap{}
	flag.Var(&lifecycles, "lifecycle", "app lifecycle binary bundle mapping (lifecycle[/stack]:bundle-filepath-in-fileserver)")
	flag.Parse()

	logger, reconfigurableSink := cflager.New("metrics_collector")
	//initializeDropsonde(logger)

	//======================= Save Process ID to .pid file ==================
	pid := os.Getpid()
	logger.Info("##### process id :", lager.Data{"process_id ":pid})

	_, err := os.Stat(*pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Fatal("Target PID File does not exist.", err)

			//Create new PID File if not exists.
			f, err := os.OpenFile(*pidFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
			defer f.Close()
			if err != nil {
				logger.Fatal("#Main: failt to create pid file.", err)
			}
			f.WriteString(strconv.Itoa(pid))
		}
	}
	err = ioutil.WriteFile(*pidFile, []byte(strconv.Itoa(pid)), 0666)
	if err != nil {
		logger.Fatal("#Main: Taget PID FIle write error :", err)
	}
	//=======================================================================

	var startTime time.Time
	//============================================
	// Get token
	cf_token, err :=  util.GetCFToken(logger, *uaaUrl, *clientId, *clientPass)
	if err != nil || cf_token == "" {
		logger.Info("#Main: There is an error hannpend getting user token", )
		os.Exit(0)
	}
	//============================================

	//============================================
	//Influx Configuration
	influxCon := new(services.InfluxConfig)
	influxCon.InfluxUrl = *influxUrl
	influxCon.InfluxDatabase = *influxDatabase
	influxCon.CfMeasurement = *cfMeasurement
	influxCon.CfProcessMeasurement = *cfProcessMeasurement

	logger.Debug("##### main.go", lager.Data{"clientId": *clientId})
	logger.Debug("##### main.go", lager.Data{"clientPass": *clientPass})
	logger.Debug("##### main.go", lager.Data{"influxUrl": *influxUrl})
	logger.Debug("##### main.go", lager.Data{"InfluxDatabase": *influxDatabase})
	logger.Debug("##### main.go", lager.Data{"cfMeasurement": *cfMeasurement})
	logger.Debug("##### main.go", lager.Data{"cfProcessMeasurement": *cfProcessMeasurement})
	logger.Debug("##### main.go", lager.Data{"dopplerUrl": *dopplerUrl})
	//============================================
	dopplerArray := strings.Split(*dopplerUrl, ",")

	members := grouper.Members{
		{"metrics_collector", handler.New(logger, dopplerArray, cf_token, *uaaUrl, *clientId, *clientPass, influxCon)},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	logger.Info("#metrics_collector started")

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	monit_err := <-monitor.Wait()

	if monit_err != nil {
		logger.Fatal("#Main: exited-with-failure", monit_err)
	}

	elapsed := time.Since(startTime)
	logger.Info("#ElapsedTime in seconds:", map[string]interface{}{"elapsed_time": elapsed, })
	logger.Info("#metrics_collector exited")
}

/*func initializeDropsonde(logger lager.Logger) {
	err := dropsonde.Initialize(dropsondeDestination, dropsondeOrigin)
	if err != nil {
		logger.Error("Main: failed to initialize dropsonde: %v", err)
	}
}*/
