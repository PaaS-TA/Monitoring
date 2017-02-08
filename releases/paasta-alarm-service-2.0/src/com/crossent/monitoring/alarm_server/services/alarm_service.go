package services

import (
	"time"
	"fmt"
	"net/http"
	"strconv"
	"net/smtp"
	"crypto/tls"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/gogobosh"
)

const (
	checkInterval   = 30 * time.Second
)

type Mail struct {
	Sender  string
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

type SmtpServer struct {
	Host      string
	Port      string
	TlsConfig *tls.Config
}

type BoshDeployments struct{
	Name 		string		`json:"deployment_name"`
	VMS 		[]gogobosh.VM	`json:"vms"`
}

type WarningMessage struct {
	ServiceName 	string
	Message 	[]string
}

type AlarmConfig struct {
	Logger          lager.Logger
	Bosh_url 	string
	Bosh_user 	string
	Bosh_pass	string
	Manager 	string
	Mgr_email 	string
	Cpu_threshold 	float64
	Mem_threshold 	float64
	Disk_threshold 	float64
	Sender 		string
	Sender_pass 	string
}

type AlarmService struct {
	gogobosh_config 		*gogobosh.Config
	alarm_config 			*AlarmConfig
}

func NewAlarmService(alarmconfig *AlarmConfig) *AlarmService {
	config := &gogobosh.Config{
		BOSHAddress: 	   alarmconfig.Bosh_url,
		Username:    	   alarmconfig.Bosh_user,
		Password:    	   alarmconfig.Bosh_pass,
		HttpClient:        http.DefaultClient,
		SkipSslValidation: true,
	}

	return &AlarmService{
		alarm_config:	alarmconfig,
		gogobosh_config: config,
	}
}

func (f AlarmService) CheckSystemThreshold() error {
	var err error
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err = f.backend_checker()
			if err != nil {
				fmt.Println("#alaram_service.CheckSystemThreshold  : There is an error during checking bosh services system metrics:", err)
			}
		}//end select
	}
	return err
}

func (f AlarmService) backend_checker() error {
	var alarmMessageArray []WarningMessage

	c, _ := gogobosh.NewClient(f.gogobosh_config)
	deployments, err := c.GetDeployments()
	if err != nil {
		f.alarm_config.Logger.Error("##### alarm_service.go - backend_checker - Get Deployments error :", err)
		return err
	}

	//var returnValue []BoshDeployments

	for _, dep := range deployments{
		boshdeployment := BoshDeployments{}
		boshdeployment.Name = dep.Name

		//fmt.Println("deployment name :", dep.Name)
		vms, err := c.GetDeploymentVMs(dep.Name)

		boshdeployment.VMS = vms
		if err != nil {
			f.alarm_config.Logger.Error("##### alarm_service.go - backend_checker - GET VM Vital info of Deployment name - Error:", err)
			return err
		}
		for _, vm :=range vms {
			var warningMsg WarningMessage
			var message_array []string

			/*fmt.Println("---------------------------------------------------------------")
			fmt.Println("id :", vm.VMID)
			fmt.Println("jobname :", vm.JobName)
			fmt.Println("index :", vm.Index)
			fmt.Println("agent_id :", vm.AgentID)
			fmt.Println("cid :", vm.CID)
			fmt.Println("dns :", vm.DNS)
			fmt.Println("vm_cid :", vm.VMCID)
			fmt.Println("vm_type :", vm.VMType)
			fmt.Println("state : ", vm.JobStatus)
			fmt.Println("cpu_sys :", vm.Vitals.CPU.Sys)
			fmt.Println("cpu_user :", vm.Vitals.CPU.User)
			fmt.Println("cpu_wait :", vm.Vitals.CPU.Wait)
			fmt.Println("mem_percent :", vm.Vitals.Mem.Percent)
			fmt.Println("mem_kb :", vm.Vitals.Mem.KB)
			fmt.Println("swap_percent :", vm.Vitals.Swap.Percent)
			fmt.Println("swap_kb :", vm.Vitals.Swap.KB)
			fmt.Println("disk system :", vm.Vitals.Disk.System.Percent)
			fmt.Println("disk ephemeral :", vm.Vitals.Disk.Ephemeral.Percent)
			fmt.Println("disk persist :", vm.Vitals.Disk.Persistent.Percent)
			//fmt.Println("### (cpu_sys + cpu_user + cpu_wait)/3, cpu_threshold :", (cpu_sys + cpu_user + cpu_wait)/3, f.alarm_config.Cpu_threshold)
			//fmt.Println("### mem_percent, mem_threshold :", mem_percent, f.alarm_config.Mem_threshold)
			//fmt.Println("### disk_system, disk_threshold :", disk_system, f.alarm_config.Disk_threshold)
			fmt.Println("---------------------------------------------------------------")*/

			//=== CPU Average ===
			cpu_sys, _ := strconv.ParseFloat(vm.Vitals.CPU.Sys, 64)
			cpu_user, _ := strconv.ParseFloat(vm.Vitals.CPU.User, 64)
			cpu_wait, _ := strconv.ParseFloat(vm.Vitals.CPU.Wait, 64)

			//=== Memory Average ===
			mem_percent, _ := strconv.ParseFloat(vm.Vitals.Mem.Percent, 64)

			//=== Disk Average ===
			disk_system, _ := strconv.ParseFloat(vm.Vitals.Disk.System.Percent, 64)

			if  (cpu_sys + cpu_user + cpu_wait)/3 > f.alarm_config.Cpu_threshold {
				warningMsg.ServiceName = vm.JobName
				message_array = append(message_array, fmt.Sprintf("CPU exceeds the previously set threshold(%.2f%%) - %.2f%%  \n", f.alarm_config.Cpu_threshold, (cpu_sys + cpu_user + cpu_wait)/3))
			}

			if mem_percent > f.alarm_config.Mem_threshold {
				warningMsg.ServiceName = vm.JobName
				message_array = append(message_array, fmt.Sprintf("Memory exceeds the previously set threshold(%.2f%%) - %.2f%% \n", f.alarm_config.Mem_threshold, mem_percent))
			}

			if disk_system > f.alarm_config.Disk_threshold {
				warningMsg.ServiceName = vm.JobName
				message_array = append(message_array, fmt.Sprintf("Disk exceeds the previously set threshold(%.2f%%) - %.2f%% \n", f.alarm_config.Disk_threshold, disk_system))
			}

			if len(message_array) > 0 {
				warningMsg.Message = message_array
				alarmMessageArray = append(alarmMessageArray, warningMsg)
			}
		}

		//fmt.Println("bosh-deployments:", boshdeployment)

		//returnValue = append(returnValue, boshdeployment)
	}

	if len(alarmMessageArray) > 0 {
		//fmt.Println("Alarm Messsage :", alarmMessageArray)
		f.SendMail(fmt.Sprintf("%v", alarmMessageArray))
	}
	return nil
	//jsonString, _ := json.Marshal(returnValue)

}

func (f AlarmService) SendMail(body string) {
	mail := Mail{}
	mail.Sender = f.alarm_config.Sender //"alarmservice.monitoring@gmail.com"
	mail.To = []string{f.alarm_config.Mgr_email}
	mail.Subject = "Warning: Some of Bosh Services reaches System threshold."
	mail.Body = body

	messageBody := f.BuildMessage(mail)

	smtpServer := SmtpServer{Host: "smtp.gmail.com", Port: "465"}
	smtpServer.TlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.Host,
	}

	auth := smtp.PlainAuth("", mail.Sender, f.alarm_config.Sender_pass, smtpServer.Host)

	conn, err := tls.Dial("tcp", smtpServer.Host + ":" + smtpServer.Port, smtpServer.TlsConfig)
	if err != nil {
		f.alarm_config.Logger.Error("smtp connection error :", err)
	}

	client, err := smtp.NewClient(conn, smtpServer.Host)
	defer client.Close()
	if err != nil {
		f.alarm_config.Logger.Error("smtp new clinet create error :", err)
	}

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		f.alarm_config.Logger.Error("client auth error :", err)
	}

	// step 2: add all from and to
	if err = client.Mail(mail.Sender); err != nil {
		f.alarm_config.Logger.Error("client send mail error :", err)
	}
	receivers := append(mail.To, mail.Cc...)
	receivers = append(receivers, mail.Bcc...)
	for _, k := range receivers {
		//fmt.Println("sending to: ", k)
		if err = client.Rcpt(k); err != nil {
			f.alarm_config.Logger.Error("sending error :", err)
		}
	}

	// Data
	w, err := client.Data()
	if err != nil {
		f.alarm_config.Logger.Error("client send data error :", err)
	}

	_, err = w.Write([]byte(messageBody))
	if err != nil {
		f.alarm_config.Logger.Error("write message error :", err)
	}

	err = w.Close()
	if err != nil {
		f.alarm_config.Logger.Error("client close error:", err)
	}

	client.Quit()
	//fmt.Println("Mail sent successfully")
}

func (f *AlarmService) BuildMessage(mail Mail) string {
	header := ""
	header += fmt.Sprintf("From: %s\r\n", mail.Sender)
	if len(mail.To) > 0 {
		header += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	}

	header += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	header += "\r\n" + mail.Body

	return header
}

/*
func (f AlarmService) SendMail(bodyMessage string) {
	// Creates a oauth2.Config using the secret
	// The second parameter is the scope, in this case we only want to send email
	conf, err := google.ConfigFromJSON(f.alarm_config.Mail_secret, gmail.GmailSendScope)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Creates a URL for the user to follow
	url := conf.AuthCodeURL("CSRF", oauth2.AccessTypeOffline)
	// Prints the URL to the terminal
	fmt.Printf("Visit this URL: \n %v \n", url)

	// Grabs the authorization code you paste into the terminal
	var code string
	_, err = fmt.Scan(&code)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Exchange the auth code for an access token
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Create the *http.Client using the access token
	client := conf.Client(oauth2.NoContext, tok)Config

	// Create a new gmail service using the client
	gmailService, err := gmail.New(client)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	// Message for our gmail service to send
	var message gmail.Message

	msg := 	[]byte(
		"From: Alarm Service \n" +
		"To: " + f.alarm_config.Mgr_email + "\n" +
		"Subject: Warning: Some of Bosh Services reaches System threshold. \n" + bodyMessage)

	// Place messageStr into message.Raw in base64 encoded format
	message.Raw = base64.URLEncoding.EncodeToString(msg)
	// Send the message
	_, err = gmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		fmt.Printf("Error: %v", err)
	} else {
		fmt.Println("Message sent!")
	}
}*/
