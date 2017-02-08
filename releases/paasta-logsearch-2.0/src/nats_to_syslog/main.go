package main

import (
	"encoding/json"
	"flag"
	"github.com/nats-io/nats"
	"github.com/pivotal-golang/lager"
	"log/syslog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type logEntry struct {
	Data    string
	Reply   string
	Subject string
}

var stop chan bool
var logger lager.Logger

func main() {
	logger = lager.NewLogger("nats-to-syslog")
	stop = make(chan bool)
	buffer := make(chan *nats.Msg, 1000)

	trapSignals()

	var natsUri = flag.String("nats-uri", "nats://localhost:4222", "The NATS server URI")
	var natsSubject = flag.String("nats-subject", ">", "The NATS subject to subscribe to")
	var syslogEndpoint = flag.String("syslog-endpoint", "localhost:514", "The remote syslog server host:port")
	var debug = flag.Bool("debug", false, "debug logging true/false")
	flag.Parse()

	setupLogger(*debug)

	syslog := connectToSyslog(*syslogEndpoint)
	defer syslog.Close()

	natsClient := connectToNATS(*natsUri)
	defer natsClient.Close()

	go func() {
		for message := range buffer {
			sendToSyslog(message, syslog)
		}
	}()

	natsClient.Subscribe(*natsSubject, func(message *nats.Msg) {
		buffer <- message
	})
	logger.Info("subscribed-to-subject", lager.Data{"subject": *natsSubject})

	<-stop
	logger.Info("bye.")
}

func handleError(err error, context string) {
	if err != nil {
		context = strings.Replace(context, " ", "-", -1)
		errorLogger := logger.Session(context)
		errorLogger.Error("error", err)
		os.Exit(1)
	}
}

func buildLogMessage(message *nats.Msg) string {
	entry := logEntry{
		Data:    string(message.Data),
		Reply:   message.Reply,
		Subject: message.Subject,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		logger.Error("unmarshalling-log-failed", err, lager.Data{"data": string(message.Data)})
		return ""
	}

	return string(data)
}

func connectToSyslog(endpoint string) *syslog.Writer {
	syslog, err := syslog.Dial("tcp", endpoint, syslog.LOG_INFO, "nats-to-syslog")
	handleError(err, "connecting to syslog")
	logger.Info("connected-to-syslog", lager.Data{"endpoint": endpoint})
	return syslog
}

func connectToNATS(natsUri string) *nats.Conn {
	natsClient, err := nats.Connect(natsUri)
	handleError(err, "connecting to nats")
	logger.Info("connected-to-nats", lager.Data{"uri": natsUri})
	return natsClient
}

func sendToSyslog(message *nats.Msg, syslog *syslog.Writer) {
	logMessage := buildLogMessage(message)
	logger.Debug("message-sent-to-syslog", lager.Data{"message": logMessage})
	err := syslog.Info(logMessage)
	if err != nil {
		logger.Error("logging-to-syslog-failed", err)
		stop <- true
	}
}

func setupLogger(debug bool) {
	if debug {
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	} else {
		logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
	}
}

func trapSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	signal.Notify(signals, syscall.SIGKILL)
	signal.Notify(signals, syscall.SIGTERM)

	go func() {
		for signal := range signals {
			logger.Info("signal-caught", lager.Data{"signal": signal})
			stop <- true
		}
	}()
}
