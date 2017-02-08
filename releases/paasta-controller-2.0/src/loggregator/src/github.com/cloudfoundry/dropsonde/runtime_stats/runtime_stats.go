package runtime_stats

import (
	"log"
	"runtime"
	"time"
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
)

type EventEmitter interface {
	Emit(events.Event) error
}

type RuntimeStats struct {
	emitter  EventEmitter
	interval time.Duration
}

func NewRuntimeStats(emitter EventEmitter, interval time.Duration) *RuntimeStats {
	return &RuntimeStats{
		emitter:  emitter,
		interval: interval,
	}
}

func (rs *RuntimeStats) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()
	for {
		rs.emit("numCPUS", float64(runtime.NumCPU()))
		rs.emit("numGoRoutines", float64(runtime.NumGoroutine()))
		rs.emitMemMetrics()

		//Add CPU Metrics
		rs.emitCpuMetrics()
		//Add Disk Metrics
		rs.emitDiskMetrics()

		select {
		case <-ticker.C:
		case <-stopChan:
			return
		}
	}
}

func (rs *RuntimeStats) emitMemMetrics() {
	/*stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	rs.emit("memoryStats.numBytesAllocatedHeap", float64(stats.HeapAlloc))
	rs.emit("memoryStats.numBytesAllocatedStack", float64(stats.StackInuse))
	rs.emit("memoryStats.numBytesAllocated", float64(stats.Alloc))
	rs.emit("memoryStats.numMallocs", float64(stats.Mallocs))
	rs.emit("memoryStats.numFrees", float64(stats.Frees))
	rs.emit("memoryStats.lastGCPauseTimeNS", float64(stats.PauseNs[(stats.NumGC+255)%256]))*/

	m, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("MemStats: failed to emit: %v", err)
	}
	rs.emit("memoryStats.TotalMemory", float64(m.Total))
	rs.emit("memoryStats.AvailableMemory", float64(m.Available))
	rs.emit("memoryStats.UsedMemory", float64(m.Used))
	rs.emit("memoryStats.UsedPercent", float64(m.UsedPercent))
}

func (rs *RuntimeStats) emit(name string, value float64) {
	err := rs.emitter.Emit(&events.ValueMetric{
		Name:  &name,
		Value: &value,
		Unit:  proto.String("count"),
	})
	if err != nil {
		log.Printf("RuntimeStats: failed to emit: %v", err)
	}
}

/*
 Description: VM - CPU Info metrics
 */
func (rs *RuntimeStats) emitCpuMetrics() {
	numcpu := runtime.NumCPU()
	duration := time.Duration(1) * time.Second
	c, err := cpu.Percent(duration, true)
	if err != nil {
		log.Println("getting cpu metrics error %v", err.Error())
		//log.Fatalf("getting cpu metrics error %v", err)
		return
	}

	for k, percent := range c {
		// Check for slightly greater then 100% to account for any rounding issues.
		if percent < 0.0 || percent > 100.0001 * float64(numcpu) {
			log.Println("CPUPercent value is invalid: %f", percent)
			//log.Fatalf("CPUPercent value is invalid: %f", percent)
		}else{
			rs.emit(fmt.Sprintf("cpuStats.%d", k), float64(percent))
		}
		//log.Println("%d cpu %f", k, percent)
	}

	//============ CPU Load Average : Only support linux & freebsd ==============
	h, err := host.Info()
	if h.OS == "linux" || h.OS == "freebsd"{
		loadAvgStat, err := load.Avg()
		if err != nil {
			log.Printf("LoadAvgStats: failed to emit: %v", err)
		}
		rs.emit("loadavg1.", float64(loadAvgStat.Load1))
		rs.emit("loadavg5.", float64(loadAvgStat.Load5))
		rs.emit("loadavg15.", float64(loadAvgStat.Load15))
	}
	//===========================================================================

}

/*
 Description: VM - Disk/IO Info metrics
 */
func (rs *RuntimeStats) emitDiskMetrics() {
	if runtime.GOOS == "windows" {
		var pathKey []string
		diskios, _ := disk.IOCounters()
		for key, _ := range diskios{
			pathKey = append(pathKey, key)
		}

		for _, value := range pathKey {
			d, err := disk.Usage(value)
			if err != nil {
				log.Println("getting disk info error %v", err.Error())
				//log.Fatalf("getting disk info error %v", err)
				return
			}

			rs.emit(fmt.Sprintf("diskStats.windows.%s.Total", d.Path), float64(d.Total))
			rs.emit(fmt.Sprintf("diskStats.windows.%s.Used", d.Path), float64(d.Used))
			rs.emit(fmt.Sprintf("diskStats.windows.%s.Available", d.Path), float64(d.Free))
			rs.emit(fmt.Sprintf("diskStats.windows.%s.Usage", d.Path), float64(d.UsedPercent))
		}

	}else{
		path := "/"

		d, err := disk.Usage(path)
		if err != nil {
			log.Println("getting disk info error %v", err.Error())
			//log.Fatalf("getting disk info error %v", err)
			return
		}
		//log.Printf("path : %s, fstype : %s, total : %d, used : %d, avail : %d, usage : %f", d.Path, d.Fstype, d.Total, d.Used, d.Free, d.UsedPercent)
		rs.emit("diskStats.Total", float64(d.Total))
		rs.emit("diskStats.Used", float64(d.Used))
		rs.emit("diskStats.Available", float64(d.Free))
		rs.emit("diskStats.Usage", float64(d.UsedPercent))
	}
}