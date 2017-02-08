package main

import (
	"os"
	"os/signal"
	"syscall"
	"strconv"
	"flag"
	"io/ioutil"
	"net"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry-incubator/runtime-schema/cc_messages/flags"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/sigmon"
	"github.com/tedsuo/ifrit/grouper"

	"com/crossent/monitoring/alarm_server/handlers"
	"com/crossent/monitoring/alarm_server/services"
)

var listenAddress = flag.String(
	"listenAddress",
	"",
	"The host:port that the server is bound to.",
)

var bosh_url = flag.String(
	"bosh_url",
	"",
	"Bosh target URL(Host:Port)",
)

var bosh_user = flag.String(
	"bosh_user",
	"",
	"Bosh Admin ID",
)

var bosh_pass = flag.String(
	"bosh_pass",
	"",
	"Bosh Admin Password",
)

var manager = flag.String(
	"manager",
	"",
	"System Manager's Name",
)

var manager_email = flag.String(
	"manager_email",
	"",
	"System Manager's Email Address",
)

var cpu_threshold = flag.String(
	"cpu_threshold",
	"",
	"Bosh Services' CPU threshold",
)


var mem_threshold = flag.String(
	"memory_threshold",
	"",
	"Bosh Services' Memory threshold",
)


var disk_threshold = flag.String(
	"disk_threshold",
	"",
	"Bosh Services's Disk threshold",
)

var sender = flag.String(
	"sender",
	"",
	"Alarm service sender",
)
var sender_pass = flag.String(
	"sender_pass",
	"",
	"Alarm service Sender's Password",
)



var pidFile = flag.String(
	"pidFile",
	"",
	"File for Current Process ID",
)

type Config map[string]string

func main(){
	//============================================
	debugserver.AddFlags(flag.CommandLine)
	cflager.AddFlags(flag.CommandLine)

	lifecycles := flags.LifecycleMap{}
	flag.Var(&lifecycles, "lifecycle", "app lifecycle binary bundle mapping (lifecycle[/stack]:bundle-filepath-in-fileserver)")
	flag.Parse()

	logger, reconfigurableSink := cflager.New("alaram_server")
	//============================================

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

	_, portString, err := net.SplitHostPort(*listenAddress)
	if err != nil {
		logger.Fatal("failed-invalid-listen-address", err)
	}
	portNum, err := net.LookupPort("tcp", portString)
	if err != nil {
		logger.Fatal("failed-invalid-listen-port", err, lager.Data{"portnum":portNum})
	}


	//============================================
	// Channel for Singal Checkig
	sigs := make(chan os.Signal, 1)
	//Waiting to be notified
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Info("returned signal:", lager.Data{"signal":sig})
		//When unexpected signal happens, defer function doesn't work.
		//So, go func has a role to be notified signal and do defer function execute
		os.Exit(0)
	}()
	//============================================

	//============================================
	// Convert String to Float64
	cpu_limit, _ := strconv.ParseFloat(*cpu_threshold, 64)
	mem_limit, _ := strconv.ParseFloat(*mem_threshold, 64)
	disk_limit, _ := strconv.ParseFloat(*disk_threshold, 64)

	// Configuration related Alarm policy
	alarm_config := new(services.AlarmConfig)
	alarm_config.Logger = logger
	alarm_config.Bosh_url = *bosh_url
	alarm_config.Bosh_user = *bosh_user
	alarm_config.Bosh_pass = *bosh_pass
	alarm_config.Manager = *manager
	alarm_config.Mgr_email = *manager_email
	alarm_config.Cpu_threshold = cpu_limit
	alarm_config.Mem_threshold = mem_limit
	alarm_config.Disk_threshold = disk_limit
	alarm_config.Sender = *sender
	alarm_config.Sender_pass = *sender_pass

	// Route Path 정보와 처리 서비스 연결
	handler := handlers.NewHandler(alarm_config)
	//============================================
	logger.Info("##### Alaram Service Server starting!!!")
	members := grouper.Members{
		{"server", handler},
	}

	if dbgAddr := debugserver.DebugAddress(flag.CommandLine); dbgAddr != "" {
		members = append(grouper.Members{
			{"debug-server", debugserver.Runner(dbgAddr, reconfigurableSink)},
		}, members...)
	}

	group := grouper.NewOrdered(os.Interrupt, members)
	monitor := ifrit.Invoke(sigmon.New(group))
	logger.Info("started")
	monit_err := <-monitor.Wait()
	if monit_err != nil {
		logger.Fatal("#Main: exited-with-failure", monit_err)
	}
	logger.Info("exited")
}
