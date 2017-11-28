 package main

 import (
	 "flag"
	 "net"
	 "os"

	 "code.cloudfoundry.org/cflager"
	 "code.cloudfoundry.org/debugserver"

	 "github.com/tedsuo/ifrit"
	 "github.com/tedsuo/ifrit/sigmon"
	 "github.com/tedsuo/ifrit/grouper"
	 "code.cloudfoundry.org/runtimeschema/cc_messages/flags"

	 "com/crossent/monitoring_agent/handler"
	 "com/crossent/monitoring_agent/services"
 )

 var origin = flag.String(
	 "origin",
	 "",
	 "VM name",
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

 var measurement = flag.String(
	 "measurement",
	 "",
	 "Influx Metrics Measurement name",
 )

 var processMeasurement = flag.String(
	 "processMeasurement",
	 "",
	 "Influx Bosh Process Measurement name",
 )

 /*var pidFile = flag.String(
	 "pidFile",
	 "",
	 "File for Current Process ID",
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
	/*pid := os.Getpid()
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
	}*/
	//=======================================================================

	//============================================
	//Influx Configuration
	influxCon := new(services.InfluxConfig)
	influxCon.InfluxUrl = *influxUrl
	influxCon.InfluxDatabase = *influxDatabase
	influxCon.Measurement = *measurement
	influxCon.ProcessMeasurement = *processMeasurement

	var cellIp string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		*origin = "127.0.0.1"
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				cellIp = ipnet.IP.String() //fmt.Println(ipnet.IP.String())
			}
		}
	}

	members := grouper.Members{
		{"metrics_sender", handler.New(logger, influxCon, *origin, cellIp)},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	logger.Info("#metrics_sender started")

	group := grouper.NewOrdered(os.Interrupt, members)

	monitor := ifrit.Invoke(sigmon.New(group))

	monit_err := <-monitor.Wait()

	if monit_err != nil {
		logger.Fatal("#Main: exited-with-failure", monit_err)
	}
}
