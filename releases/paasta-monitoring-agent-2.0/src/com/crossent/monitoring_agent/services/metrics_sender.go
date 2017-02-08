package services

import (
	"fmt"
	"time"
	"runtime"

	"code.cloudfoundry.org/lager"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/disk"
)

const (
	statsInterval        = 30 * time.Second
)
type MetricSender struct {
	logger 		lager.Logger
	influx 		*InfluxConfig
	origin 		string
	cellIp 		string
}

type InfluxConfig struct {
	InfluxUrl		string
	InfluxUser 		string
	InfluxPass 		string
	InfluxDatabase 		string
	Measurement 		string
}

type CpuStats struct{
	CoreNum		int
	Status 		float64
}

type MemStatus struct{
	Total 		float64
	Available 	float64
	Used 		float64
	UsedPercent 	float64
}

func NewMetricSender(logger lager.Logger, influx *InfluxConfig, origin, cellIp string) *MetricSender{
	return &MetricSender{
		logger:		logger,
		influx:		influx,
		origin: 	origin,
		cellIp: 	cellIp,
	}
}
func (f *MetricSender) SendMetricsToInfluxDb() error {
	f.logger.Info("influx :", map[string]interface{}{"influxUrl":f.influx.InfluxUrl, "influxdatabase":f.influx.InfluxDatabase})

	var err error
	ticker := time.NewTicker(statsInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err = f.Collect_Save()
			if err != nil {
				f.logger.Error("#metrics_sender.SendMetricsToInfluxDb  : There is an error during sending metrics to influxdb:", err)
			}
		}//end select
	}
	return err
}

func (f *MetricSender) Collect_Save() error {
	//===============================================================
	//catch or finally
	/*defer func() {
		if err := recover(); err != nil { //catch
			//fmt.Fprintf(os.Stderr, "Exception at metrics_sender.Collect_Save(): %v\n", err)
			//os.Exit(1)
			return err
		}
	}()*/
	//===============================================================
	// Make client
	c, err := influxdb.NewUDPClient(influxdb.UDPConfig{
		Addr: f.influx.InfluxUrl,
		//PayloadSize: 4096,
	})

	if err != nil {
		return err
	}

	// Create a new point batch
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  f.influx.InfluxDatabase,
		Precision: "s",
	})

	if err != nil {
		return err
	}

	//======================================================================
	// ========= CPU Status ========
	//======================================================================
	cpuStatus, load1avg, load5avg, load15avg := f.emitCpuMetrics()
	for _, v := range cpuStatus{
		//points = append(points, f.makePoint(f.origin, f.cellIp, fmt.Sprintf("cpuStats.core.%d", v.CoreNum), float64(v.Status)))
		//fmt.Println("core number : ", v.CoreNum)
		//fmt.Println("core status : ", v.Status)
		bp.AddPoint(f.makePoint(f.origin, f.cellIp, fmt.Sprintf("cpuStats.core.%d", v.CoreNum), float64(v.Status)))
	}

	//points = append(points, f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg1Stats", float64(load1avg)))
	//points = append(points, f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg5Stats", float64(load5avg)))
	//points = append(points, f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg15Stats", float64(load15avg)))

	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg1Stats", float64(load1avg)))
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg5Stats", float64(load5avg)))
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "cpuStats.LoadAvg15Stats", float64(load15avg)))
	/*fmt.Println("load 1 avg : ", load1avg)
	fmt.Println("load 5 avg : ", load5avg)
	fmt.Println("load 15 avg : ", load15avg)*/
	//======================================================================

	//======================================================================
	// ========= Memory Status ========
	//======================================================================
	m, err := mem.VirtualMemory()
	/*fmt.Println("memoryStats.TotalMemory", float64(m.Total))
	fmt.Println("memoryStats.AvailableMemory", float64(m.Available))
	fmt.Println("memoryStats.UsedMemory", float64(m.Used))
	fmt.Println("memoryStats.UsedPercent", float64(m.UsedPercent))*/

	/*points = append(points, f.makePoint(f.origin, f.cellIp, "memoryStats.TotalMemory", float64(m.Total)))
	points = append(points, f.makePoint(f.origin, f.cellIp, "memoryStats.AvailableMemory", float64(m.Available)))
	points = append(points, f.makePoint(f.origin, f.cellIp, "memoryStats.UsedMemory", float64(m.Used)))
	points = append(points, f.makePoint(f.origin, f.cellIp, "memoryStats.UsedPercent", float64(m.UsedPercent)))*/
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "memoryStats.TotalMemory", float64(m.Total)))
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "memoryStats.AvailableMemory", float64(m.Available)))
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "memoryStats.UsedMemory", float64(m.Used)))
	bp.AddPoint(f.makePoint(f.origin, f.cellIp, "memoryStats.UsedPercent", float64(m.UsedPercent)))

	if err != nil {
		f.logger.Error("MemStats: failed to emit: %v", err)
	}
	//======================================================================


	//======================================================================
	// ========= Disk Status ========
	//======================================================================
	if runtime.GOOS == "windows" {
		var pathKey []string
		diskios, _ := disk.IOCounters()
		for key, _ := range diskios{
			pathKey = append(pathKey, key)
		}
		for _, value := range pathKey {
			d, err := disk.Usage(value)
			if err != nil {
				f.logger.Error("getting disk info error %v", err)
			}
			/*fmt.Println(fmt.Sprintf("diskStats.windows.%s.Total", d.Path), float64(d.Total))
			fmt.Println(fmt.Sprintf("diskStats.windows.%s.Used", d.Path), float64(d.Used))
			fmt.Println(fmt.Sprintf("diskStats.windows.%s.Available", d.Path), float64(d.Free))
			fmt.Println(fmt.Sprintf("diskStats.windows.%s.Usage", d.Path), float64(d.UsedPercent))*/

			/*points = append(points, f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Total", d.Path), float64(d.Total)))
			points = append(points, f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Used", d.Path), float64(d.Used)))
			points = append(points, f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Available", d.Path), float64(d.Free)))
			points = append(points, f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Usage", d.Path), float64(d.UsedPercent)))*/
			bp.AddPoint(f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Total", d.Path), float64(d.Total)))
			bp.AddPoint(f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Used", d.Path), float64(d.Used)))
			bp.AddPoint(f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Available", d.Path), float64(d.Free)))
			bp.AddPoint(f.makePoint(f.origin, f.cellIp, fmt.Sprintf("diskStats.windows.%s.Usage", d.Path), float64(d.UsedPercent)))
		}
	}else{
		path := "/"
		d, err := disk.Usage(path)
		if err != nil {
			f.logger.Error("getting disk info error %v", err)
		}
		//fmt.Println("diskStats.Total", float64(d.Total))
		//fmt.Println("diskStats.Used", float64(d.Used))
		//fmt.Println("diskStats.Available", float64(d.Free))
		//fmt.Println("diskStats.Usage", float64(d.UsedPercent))

		/*points = append(points, f.makePoint(f.origin, f.cellIp, "diskStats.Total", float64(d.Total)))
		points = append(points, f.makePoint(f.origin, f.cellIp, "diskStats.Used", float64(d.Used)))
		points = append(points, f.makePoint(f.origin, f.cellIp, "diskStats.Available", float64(d.Free)))
		points = append(points, f.makePoint(f.origin, f.cellIp, "diskStats.Usage", float64(d.UsedPercent)))*/

		bp.AddPoint(f.makePoint(f.origin, f.cellIp, "diskStats.Total", float64(d.Total)))
		bp.AddPoint(f.makePoint(f.origin, f.cellIp, "diskStats.Used", float64(d.Used)))
		bp.AddPoint(f.makePoint(f.origin, f.cellIp, "diskStats.Available", float64(d.Free)))
		bp.AddPoint(f.makePoint(f.origin, f.cellIp, "diskStats.Usage", float64(d.UsedPercent)))
	}

	// Write the batch
	c.Write(bp)
	c.Close()
	return nil
}

// Creates a measurement point with a single value field
func (f *MetricSender) makePoint(origin, cellIp, name string, value float64) *influxdb.Point {
	var tags map[string]string
	var fields map[string]interface{}
	tags = map[string]string{
		"origin": origin,
		"ip": cellIp,
		"metricname" : name,
	}
	fields = map[string]interface{}{
		"value": value,
	}

	mkPoint, err := influxdb.NewPoint(f.influx.Measurement, tags, fields)
	if err != nil {
		f.logger.Error("#metrics_collector.makePoint : error caused during making a new point", err)
	}

	return mkPoint
}

/*
 Description: VM - CPU Info metrics
 */
func (f *MetricSender) emitCpuMetrics() ([]CpuStats, float64, float64, float64) {
	numcpu := runtime.NumCPU()
	duration := time.Duration(1) * time.Second
	c, err := cpu.Percent(duration, true)
	if err != nil {
		f.logger.Error("getting cpu metrics error %v", err)
		return nil, 0, 0, 0
	}

	var cpuStatusArray []CpuStats
	var load1AvgStat, load5AvgStat, load15AvgStat float64
	for k, percent := range c {
		var cpuStatus CpuStats
		// Check for slightly greater then 100% to account for any rounding issues.
		if percent < 0.0 || percent > 100.0001 * float64(numcpu) {
			f.logger.Info("CPUPercent value is invalid: %f", lager.Data{"percent":percent})
		}else{
			cpuStatus.CoreNum = k
			cpuStatus.Status = float64(percent)

		}
		cpuStatusArray = append(cpuStatusArray, cpuStatus)
	}

	//============ CPU Load Average : Only support linux & freebsd ==============
	h, err := host.Info()
	if h.OS == "linux" || h.OS == "freebsd"{
		loadAvgStat, err := load.Avg()
		if err != nil {
			f.logger.Error("LoadAvgStats: failed to get LoadAvg information: %v", err)
		}
		load1AvgStat = float64(loadAvgStat.Load1)
		load5AvgStat = float64(loadAvgStat.Load5)
		load15AvgStat = float64(loadAvgStat.Load15)
	}
	//===========================================================================
	return cpuStatusArray, load1AvgStat, load5AvgStat, load15AvgStat
}
