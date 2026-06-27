package collector

import (
	"runtime"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Struct datos principales
type Metrics struct {
	CPUUsage     float64         `json:"cpu_percent"`
	RAMUsedMB    float64         `json:"ram_used_mb"`
	RAMTotalMB   float64         `json:"ram_total_mb"`
	DiskUsage    []DiskPartition `json:"disk_usage"`
	NetworkSent  float64         `json:"network_out_kbps"`
	NetworkRecv  float64         `json:"network_in_kbps"`
	TopProcesses []ProcessInfo   `json:"top_processes"`
}

// Struct de una partición individual de disco
type DiskPartition struct {
	Partition string  `json:"partition"`
	UsedGB    float64 `json:"used_gb"`
	TotalGB   float64 `json:"total_gb"`
}

// Struct de un proceso individual
type ProcessInfo struct {
	Name       string  `json:"name"`
	PID        int32   `json:"pid"`
	CPUPercent float64 `json:"cpu_percent"`
}

func getCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return 0, err
	}
	return percentages[0], nil
}

func getMemoryUsage() (float64, float64, error) {
	meminfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}

	usedMB := float64(meminfo.Used) / 1024 / 1024
	totalMB := float64(meminfo.Total) / 1024 / 1024

	return usedMB, totalMB, nil
}

func getPathInfo() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}

// Lista todas las particiones reales del sistema y su uso individual
func getDiskUsage() ([]DiskPartition, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var resultado []DiskPartition

	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}

		resultado = append(resultado, DiskPartition{
			Partition: p.Mountpoint,
			UsedGB:    float64(usage.Used) / 1024 / 1024 / 1024,
			TotalGB:   float64(usage.Total) / 1024 / 1024 / 1024,
		})
	}

	return resultado, nil
}

func getNetworkUsage() (float64, float64, error) {
	counters1, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}
	time.Sleep(500 * time.Millisecond)
	counters2, err := net.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}

	bytesSentDiff := counters2[0].BytesSent - counters1[0].BytesSent
	bytesRecvDiff := counters2[0].BytesRecv - counters1[0].BytesRecv

	return float64(bytesSentDiff) * 2, float64(bytesRecvDiff) * 2, nil
}

// Lista los procesos que más CPU consumen, limitado a "limit" resultados
func getTopProcesses(limit int) ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var lista []ProcessInfo
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpuPct, err := p.CPUPercent()
		if err != nil {
			continue
		}
		lista = append(lista, ProcessInfo{
			Name:       name,
			PID:        p.Pid,
			CPUPercent: cpuPct,
		})
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].CPUPercent > lista[j].CPUPercent
	})

	if len(lista) > limit {
		lista = lista[:limit]
	}

	return lista, nil
}

func GetAllMetrics() (*Metrics, error) {
	var m Metrics

	cpuPct, err := getCPUUsage()
	if err != nil {
		return nil, err
	}
	m.CPUUsage = cpuPct

	usedMB, totalMB, err := getMemoryUsage()
	if err != nil {
		return nil, err
	}
	m.RAMUsedMB = usedMB
	m.RAMTotalMB = totalMB

	partitions, err := getDiskUsage()
	if err != nil {
		return nil, err
	}
	m.DiskUsage = partitions

	netSent, netRecv, err := getNetworkUsage()
	if err != nil {
		return nil, err
	}
	m.NetworkSent = netSent
	m.NetworkRecv = netRecv

	topProcs, err := getTopProcesses(5)
	if err != nil {
		return nil, err
	}
	m.TopProcesses = topProcs

	return &m, nil
}