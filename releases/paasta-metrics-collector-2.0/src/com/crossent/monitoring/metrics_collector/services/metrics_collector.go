package services

import (
	"time"
	"strings"
	"sync"
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/influxdata/influxdb/client/v2"
	"com/crossent/monitoring/metrics_collector/util"
	"strconv"
)

const firehoseSubscriptionId string = "firehose-prototype"

type FirehoseConsumer struct {
	logger 		lager.Logger
	consumer 	*consumer.Consumer
	token		string
	msgChan 	<-chan *events.Envelope
	errChan		<-chan error
	uaaUrl		string
	client_id 	string
	client_pass 	string
	influx 		*InfluxConfig
	retry 		bool
}

type InfluxConfig struct {
	InfluxUrl		string
	InfluxUser 		string
	InfluxPass 		string
	InfluxDatabase 		string
	CfMeasurement 		string
	CfProcessMeasurement 	string
}

func NewFiehoseConsumer(logger lager.Logger, consumer *consumer.Consumer, token, uaaUrl, client_id, client_pass string, influx *InfluxConfig) *FirehoseConsumer{
	return &FirehoseConsumer{
		logger:		logger,
		consumer:	consumer,
		token:		token,
		uaaUrl:		uaaUrl,
		client_id: 	client_id,
		client_pass: 	client_pass,
		influx:		influx,
		retry:		false,
	}
}

func (f *FirehoseConsumer) SetToken(token string) {
	f.token = token
}

func (f *FirehoseConsumer) GetMetricsStream(index int) {
	var wg sync.WaitGroup
	wg.Add(2)

	f.msgChan, f.errChan = f.consumer.Firehose(firehoseSubscriptionId, f.token)

	go func(wg *sync.WaitGroup){
		defer wg.Done()
		f.SendMetricsToInfluxDb(index)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		f.ErrorHandling(index)
	}(&wg)
	wg.Wait()
	f.logger.Debug("# metrics_collector.GetMetricsStream end ...")
}

func (f *FirehoseConsumer) SendMetricsToInfluxDb(index int)  {
	f.logger.Info("influx :", map[string]interface{}{"influxUrl":f.influx.InfluxUrl, "influxdatabase":f.influx.InfluxDatabase})

	// Make client
	c, err := client.NewUDPClient(client.UDPConfig{
		Addr: f.influx.InfluxUrl,
		//PayloadSize: 4096,
	})

	if err != nil {
		f.logger.Error("#metrics_collector.SendMetricsToInfluxDb  : There is an error during connecting influxdb to store metrics:", err)
		return
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  f.influx.InfluxDatabase,
		Precision: "s",
	})

	if err != nil {
		f.logger.Error("#metrics_collector.SendMetricsToInfluxDb : error caused during creating a new point batch",  err)
		return
	}

	var name string
	var value , delta, total, count float64
	var tags map[string]string
	var fields map[string]interface{}

	//==================================================================================================================
	//Timer : if there is no input from websocket for 10*time.Second, consider the websocket broken and restart process
	timerChan := time.NewTimer(time.Second * 10)
	go func() {
		select {
		case <-timerChan.C:
		//fmt.Println("=========================== timer working.... 10 seconds..........")
			f.logger.Error("#metrics_collector.SendMetricsToInfluxDb : There is no response from websocket & need to restart process.", errors.New("No response from websocket!!!"))
			f.consumer.Close()
			return
		}
	}()
	//==================================================================================================================
	for msg := range f.msgChan {
		//fmt.Println("message is processing...")
		//f.logger.Info("### message from channel :", lager.Data{"msg":msg})

		//=============================================================
		//convert timestamp (int64) to time object
		str_timestamp := strconv.FormatInt(int64(*msg.Timestamp),10)
		sec_time, _ := strconv.Atoi(str_timestamp[:10])
		nsec_time, _ := strconv.Atoi(str_timestamp[10:len(str_timestamp)])
		sendedtime := time.Unix(int64(sec_time), int64(nsec_time))
		//=============================================================

		//=======================================================================================
		//Origin="MetronAgent"에서 수집되는 System Metrics 정보는 다른 Origin 정보들과 중복된다.
		//Origin="MetronAgent" 이면서 job = 'api' or 'uaa' 인 경우에는 MetronAgent 정보가 필요하여, Filtering 주석처리
		//=======================================================================================
		//if *msg.Origin != "MetronAgent"{
			job_zone := strings.Split(*msg.Job, "_")

			//f.logger.Info("metrics msg", map[string]interface{}{"metrics": msg})
			if msg.ValueMetric != nil {
				//f.logger.Info("ValueMetric metrics - name & value & unit:", map[string]interface{}{"name":*msg.ValueMetric.Name, "value":strconv.FormatFloat(*msg.ValueMetric.Value, 'f', 6, 64), "unit":*msg.ValueMetric.Unit})
				name = *msg.ValueMetric.Name
				value = *msg.ValueMetric.Value //strconv.FormatFloat(*msg.ValueMetric.Value, 'f', 6, 64)
			}else if msg.CounterEvent != nil {
				//f.logger.Info("CounterEvent metrics - name & value & unit:", map[string]interface{}{"name":*msg.CounterEvent.Name, "delta":*msg.CounterEvent.Delta, "total":*msg.CounterEvent.Total})
				name = *msg.CounterEvent.Name
				delta = float64(*msg.CounterEvent.Delta)
				total = float64(*msg.CounterEvent.Total)
			}

			//if metricname starts with "processStat", save metric info into cf process measurement, else save into cf measurement.
			if strings.Contains(name, "processStats") {
				// metricname = processStats.0.metron.startTime
				//  structure = processStats."index"."name"."gubun"
				proc_array := strings.Split(name, ".")

				//Set Tags - origin, eventtype, job, zone, index, ip
				tags = map[string]string{
					"origin": *msg.Origin,
					"eventtype": msg.EventType.String(),
					"job": string(*msg.Job)[0:strings.LastIndex(*msg.Job, "_")], 	//"job": job_zone[0],
					"zone": job_zone[len(job_zone)-1],				//"zone":job_zone[1],
					"index":*msg.Index,
					"ip": *msg.Ip,
					"proc_name": proc_array[2],
					"proc_gubun": proc_array[3],
				}

				//Set Fields - metricname, proc_index, proc_name, proc_gubun, mem_usage, delta, total
				procindex, _ := strconv.Atoi(proc_array[1])
				fields = map[string]interface{}{
					"metricname": proc_array[0],
					"proc_index": procindex,
					"mem_usage": value,
					"delta": delta,
					"total": total,
				}

				pt, err := client.NewPoint(f.influx.CfProcessMeasurement, tags, fields, sendedtime) //time.Unix(*msg.Timestamp, 0))

				if err != nil {
					f.logger.Error("#metrics_collector.SendMetricsToInfluxDb : error caused during a new point batch", err)
				}

				bp.AddPoint(pt)
			}else{
				// Create a point and add to batch
				//Set Tags - origin, eventtype, job, zone, index, metricname, ip
				tags = map[string]string{
					"origin": *msg.Origin,
					"eventtype": msg.EventType.String(),
					"job": string(*msg.Job)[0:strings.LastIndex(*msg.Job, "_")], 	//"job": job_zone[0],
					"zone": job_zone[len(job_zone)-1],				//"zone":job_zone[1],
					"index":*msg.Index,
					"metricname": name,
					"ip": *msg.Ip,
				}

				//f.logger.Info("tags", lager.Data{"origin":*msg.Origin, "job":job_zone[0], "zone":job_zone[1], "metricname":name, "ip":*msg.Ip})

				//Set Fields - name, value, total
				fields = map[string]interface{}{
					//"name":   name,
					"value": value,
					"delta": delta,
					"total": total,
				}

				pt, err := client.NewPoint(f.influx.CfMeasurement, tags, fields, sendedtime) //time.Unix(*msg.Timestamp, 0))

				if err != nil {
					f.logger.Error("#metrics_collector.SendMetricsToInfluxDb : error caused during a new point batch", err)
				}

				bp.AddPoint(pt)
			}


			// Buffering til size 2000 & Save
			if count == 2000 {
				// Write the batch
				c.Write(bp)
				count = 0
			}
			value = 0.0
			delta = 0.0
			total = 0.0
			count++

			//Initialize Timer
			timerChan.Reset(time.Second*10)

		//}
		//fmt.Println("=========================== message to influx finished ==============================")
	}
	f.logger.Info("# metrics_collector.SendMetricsToInfluxDb end ...")
	return
}

func (f *FirehoseConsumer) ErrorHandling(k int) {
	//Set retryCount for reconnect firehose.
	for err := range f.errChan {
		//if unexpected Error Happened, noaa firehose retryAction called.
		f.logger.Error("# metrics_collector.ErroHandling : ", err)
		if strings.Contains(err.Error(), "Unauthorized")  {
			cf_token, err := util.GetCFToken(f.logger,f.uaaUrl,f.client_id,f.client_pass)
			if err != nil {
				f.logger.Error("# noaa_agent.ErroHandling : There is an error hannpend getting user token", err)
				//f.errChan <- errors.New("Unauthorized")
			}else {
				f.consumer.Close()
				f.SetToken(cf_token)
				f.GetMetricsStream(k)
			}
		} else {
			f.logger.Error("# noaa_agent.ErroHandling : There is an error hannpend", err)
			f.consumer.Close()
			return
		}
	}
}
