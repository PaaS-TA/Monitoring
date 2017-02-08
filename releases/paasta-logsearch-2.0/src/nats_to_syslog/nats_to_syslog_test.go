package main_test

import (
	"bufio"
	"github.com/nats-io/nats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"net"
	"os/exec"
	"time"
)

var testBinaryPath string

var _ = BeforeSuite(func() {
	var err error
	testBinaryPath, err = gexec.Build("github.com/logsearch/nats-to-syslog")
	handleError(err)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

var _ = Describe("NatsToSyslog", func() {
	var (
		gnatsd       *gexec.Session
		natsClient   *nats.Conn
		syslogServer *net.TCPListener
	)

	BeforeEach(func() {
		gnatsd = startGNATSd()

		var err error
		natsClient, err = nats.Connect("nats://nats:c1oudc0w@127.0.0.1:4567")
		handleError(err)

		syslogServer = startSyslogServer()
	})

	It("forwards NATS messages to syslog", func() {
		testBinary := exec.Command(testBinaryPath, "-nats-uri", "nats://nats:c1oudc0w@127.0.0.1:4567", "-syslog-endpoint", "localhost:6789", "-nats-subject", "testSubject", "-debug", "true")
		testSession, err := gexec.Start(testBinary, GinkgoWriter, GinkgoWriter)
		handleError(err)
		defer testSession.Kill()

		var remoteClient net.Conn
		go func() {
			var err error
			remoteClient, err = syslogServer.AcceptTCP()
			remoteClient.SetReadDeadline(time.Now().Add(2 * time.Second))
			handleError(err)
		}()
		time.Sleep(500 * time.Millisecond)
		reader := bufio.NewReader(remoteClient)
		defer remoteClient.Close()

		natsClient.Publish("testSubject", []byte("test message"))
		time.Sleep(1 * time.Second)

		logLine, _, err := reader.ReadLine()
		handleError(err)
		Expect(string(logLine)).To(MatchRegexp(`^<6>.*nats-to-syslog.*{"Data":"test message","Reply":"","Subject":"testSubject"}`))
	})

	AfterEach(func() {
		gnatsd.Kill()
		syslogServer.Close()
	})

})

func startSyslogServer() *net.TCPListener {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 6789})
	handleError(err)

	return listener
}

func startGNATSd() *gexec.Session {
	cmd := exec.Command("gnatsd", "--port", "4567", "--user", "nats", "--pass", "c1oudc0w", "-D", "-V")

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	time.Sleep(1 * time.Second)
	handleError(err)

	return session
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
