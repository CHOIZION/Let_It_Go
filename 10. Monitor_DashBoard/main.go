package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("터미널 UI 초기화 실패: %v", err)
	}
	defer ui.Close()

	cpuUsageChart := widgets.NewPlot()
	cpuUsageChart.Title = "CPU Usage"
	cpuUsageChart.Data = make([][]float64, 1)
	cpuUsageChart.Data[0] = []float64{0}
	cpuUsageChart.SetRect(0, 0, 50, 15)
	cpuUsageChart.AxesColor = ui.ColorWhite
	cpuUsageChart.LineColors[0] = ui.ColorGreen
	cpuUsageChart.DrawDirection = widgets.DrawLeft

	memUsageGauge := widgets.NewGauge()
	memUsageGauge.Title = "Memory Usage"
	memUsageGauge.SetRect(51, 0, 100, 5)
	memUsageGauge.BarColor = ui.ColorRed
	memUsageGauge.LabelStyle = ui.NewStyle(ui.ColorYellow)
	memUsageGauge.TitleStyle = ui.NewStyle(ui.ColorClear)

	diskUsageGauge := widgets.NewGauge()
	diskUsageGauge.Title = "Disk Usage"
	diskUsageGauge.SetRect(51, 6, 100, 11)
	diskUsageGauge.BarColor = ui.ColorBlue
	diskUsageGauge.LabelStyle = ui.NewStyle(ui.ColorYellow)
	diskUsageGauge.TitleStyle = ui.NewStyle(ui.ColorClear)

	netUsageChart := widgets.NewPlot()
	netUsageChart.Title = "Network Usage"
	netUsageChart.Data = make([][]float64, 2)
	netUsageChart.Data[0] = []float64{0}
	netUsageChart.Data[1] = []float64{0}
	netUsageChart.SetRect(0, 16, 100, 23)
	netUsageChart.AxesColor = ui.ColorWhite
	netUsageChart.LineColors[0] = ui.ColorYellow
	netUsageChart.LineColors[1] = ui.ColorMagenta
	netUsageChart.DrawDirection = widgets.DrawLeft

	var prevNetIOCounters []net.IOCountersStat
	prevNetIOCounters, _ = net.IOCounters(false)

	procTable := widgets.NewTable()
	procTable.Title = "Top Processes"
	procTable.SetRect(0, 24, 100, 50)
	procTable.RowSeparator = false
	procTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	procTable.Rows = [][]string{
		{"PID", "Name", "CPU%", "Memory%"},
	}

	tempTable := widgets.NewTable()
	tempTable.Title = "Temperature Sensors"
	tempTable.SetRect(0, 51, 100, 60)
	tempTable.RowSeparator = false
	tempTable.TextStyle = ui.NewStyle(ui.ColorWhite)
	tempTable.Rows = [][]string{
		{"Sensor", "Temperature (°C)"},
	}

	_, _ = cpu.Percent(0, false)
	_, _ = cpu.Percent(0, true)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	selectedProcessIndex := 0

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Down>":
				if selectedProcessIndex < len(procTable.Rows)-2 {
					selectedProcessIndex++
				}
			case "<Up>":
				if selectedProcessIndex > 0 {
					selectedProcessIndex--
				}
			case "k":
				if selectedProcessIndex > 0 {
					pidStr := procTable.Rows[selectedProcessIndex+1][0]
					pid, _ := strconv.Atoi(pidStr)
					proc, err := process.NewProcess(int32(pid))
					if err == nil {
						proc.Kill()
					}
				}
			}
		case <-ticker:
			cpuPercent, err := cpu.Percent(0, false)
			if err == nil && len(cpuPercent) > 0 {
				if len(cpuUsageChart.Data[0]) > 50 {
					cpuUsageChart.Data[0] = cpuUsageChart.Data[0][1:]
				}
				cpuUsageChart.Data[0] = append(cpuUsageChart.Data[0], cpuPercent[0])
				cpuUsageChart.Title = fmt.Sprintf("CPU Usage: %.2f%%", cpuPercent[0])
			}

			vmStat, err := mem.VirtualMemory()
			if err == nil {
				memUsageGauge.Percent = int(vmStat.UsedPercent)
				memUsageGauge.Label = fmt.Sprintf("%.2f%% (%.2fGB / %.2fGB)", vmStat.UsedPercent, byteToGB(vmStat.Used), byteToGB(vmStat.Total))
			}

			diskStat, err := disk.Usage("/")
			if err == nil {
				diskUsageGauge.Percent = int(diskStat.UsedPercent)
				diskUsageGauge.Label = fmt.Sprintf("%.2f%% (%.2fGB / %.2fGB)", diskStat.UsedPercent, byteToGB(diskStat.Used), byteToGB(diskStat.Total))
			}

			netIOCounters, err := net.IOCounters(false)
			if err == nil && len(netIOCounters) > 0 {
				bytesSent := netIOCounters[0].BytesSent - prevNetIOCounters[0].BytesSent
				bytesRecv := netIOCounters[0].BytesRecv - prevNetIOCounters[0].BytesRecv

				if len(netUsageChart.Data[0]) > 50 {
					netUsageChart.Data[0] = netUsageChart.Data[0][1:]
					netUsageChart.Data[1] = netUsageChart.Data[1][1:]
				}
				netUsageChart.Data[0] = append(netUsageChart.Data[0], float64(bytesSent))
				netUsageChart.Data[1] = append(netUsageChart.Data[1], float64(bytesRecv))
				netUsageChart.Title = fmt.Sprintf("Network Usage (Sent: %v B/s, Recv: %v B/s)", bytesSent, bytesRecv)

				prevNetIOCounters = netIOCounters
			}

			procs, err := process.Processes()
			if err == nil {
				procTable.Rows = [][]string{
					{"PID", "Name", "CPU%", "Memory%"},
				}
				var procInfos []ProcInfo
				for _, p := range procs {
					cpuPercent, _ := p.CPUPercent()
					memPercent, _ := p.MemoryPercent()
					name, _ := p.Name()
					if cpuPercent > 0.1 {
						procInfos = append(procInfos, ProcInfo{
							PID:    p.Pid,
							Name:   name,
							CPU:    cpuPercent,
							Memory: memPercent,
						})
					}
				}
				sort.Slice(procInfos, func(i, j int) bool {
					return procInfos[i].CPU > procInfos[j].CPU
				})
				for i, info := range procInfos {
					if i >= 10 {
						break
					}
					procTable.Rows = append(procTable.Rows, []string{
						fmt.Sprintf("%d", info.PID),
						info.Name,
						fmt.Sprintf("%.2f", info.CPU),
						fmt.Sprintf("%.2f", info.Memory),
					})
				}

				for i := range procTable.Rows {
					if i == selectedProcessIndex+1 {
						procTable.RowStyles[i] = ui.NewStyle(ui.ColorBlack, ui.ColorYellow, ui.ModifierBold)
					} else {
						procTable.RowStyles[i] = ui.NewStyle(ui.ColorWhite)
					}
				}
			}

			temps, err := host.SensorsTemperatures()
			if err == nil {
				tempTable.Rows = [][]string{
					{"Sensor", "Temperature (°C)"},
				}
				for _, t := range temps {
					if t.Temperature > 0 {
						tempTable.Rows = append(tempTable.Rows, []string{
							t.SensorKey,
							fmt.Sprintf("%.2f", t.Temperature),
						})
					}
				}
			} else {
				tempTable.Rows = [][]string{
					{"Temperature information not available"},
				}
			}

			hostInfo, err := host.Info()
			if err == nil {
				uptime := time.Duration(hostInfo.Uptime) * time.Second
				cpuUsageChart.Title = fmt.Sprintf("CPU Usage | Uptime: %s", uptime.String())
			}

			ui.Render(cpuUsageChart, memUsageGauge, diskUsageGauge, netUsageChart, procTable, tempTable)
		}
	}
}

type ProcInfo struct {
	PID    int32
	Name   string
	CPU    float64
	Memory float32
}

func byteToGB(b uint64) float64 {
	return float64(b) / math.Pow(1024, 3)
}
