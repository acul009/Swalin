package rmm

import (
	"fmt"
	"rahnit-rmm/util"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type StaticStats struct {
	HostInfo *host.InfoStat
}

type ActiveStats struct {
	Cpu       *CpuStats
	Memory    *MemoryStats
	Processes *ProcessStats
}

type CpuStats struct {
	Usage []float64
}

type MemoryStats struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
}

type ProcessStats struct {
	Processes []ProcessInfo
}

type ProcessInfo struct {
	Pid  int32
	Name string
}

func GetStaticStats() (*StaticStats, error) {
	hostInfo, err := GetHostInfo()
	if err != nil {
		return nil, fmt.Errorf("error retrieving host info: %w", err)
	}

	return &StaticStats{
		HostInfo: hostInfo,
	}, nil
}

func GetHostInfo() (*host.InfoStat, error) {
	return host.Info()
}

func GetActiveStats() (*ActiveStats, error) {

	memStats, err := GetMemoryStats()
	if err != nil {
		return nil, fmt.Errorf("error retrieving memory stats: %w", err)
	}

	cpuStats, err := GetCpuStats()
	if err != nil {
		return nil, fmt.Errorf("error retrieving cpu stats: %w", err)
	}

	processStats, err := GetProcessInfo()
	if err != nil {
		return nil, fmt.Errorf("error retrieving process stats: %w", err)
	}

	return &ActiveStats{
		Cpu:       cpuStats,
		Memory:    memStats,
		Processes: processStats,
	}, nil
}

func GetMemoryStats() (*MemoryStats, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("error getting virtual memory: %w", err)
	}
	return &MemoryStats{
		Total:       v.Total,
		Available:   v.Available,
		Used:        v.Used,
		UsedPercent: v.UsedPercent,
	}, nil
}

func GetCpuStats() (*CpuStats, error) {
	cpuUsage, err := cpu.Percent(0, true)
	if err != nil {
		return nil, fmt.Errorf("error retrieving cpu usage: %w", err)
	}

	return &CpuStats{
		Usage: cpuUsage,
	}, nil
}

func GetProcessInfo() (*ProcessStats, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("error getting processes: %w", err)
	}

	processesInfo := make([]ProcessInfo, 0, len(processes))

	for _, p := range processes {
		name, _ := p.Name()
		processesInfo = append(processesInfo, ProcessInfo{
			Name: name,
			Pid:  p.Pid,
		})
	}

	return &ProcessStats{Processes: processesInfo}, nil
}

func MonitorProcesses(errChan chan<- error) (util.UpdateableMap[int32, *ProcessInfo], error) {

	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("error getting processes: %w", err)
	}

	list := util.NewObservableMap[int32, *ProcessInfo]()

	for _, p := range processes {
		name, _ := p.Name()
		list.Set(p.Pid, &ProcessInfo{
			Name: name,
			Pid:  p.Pid,
		})
	}

	go func() {
		for {
			pids, err := process.Pids()
			if err != nil {
				errChan <- fmt.Errorf("error getting processes: %w", err)
				return
			}

			known := list.GetAll()

			for _, pid := range pids {
				_, ok := known[pid]
				if !ok {

					process, err := process.NewProcess(pid)
					if err != nil {
						errChan <- fmt.Errorf("error getting process: %w", err)
						return
					}

					name, _ := process.Name()

					list.Set(pid, &ProcessInfo{
						Name: name,
						Pid:  pid,
					})
				} else {
					delete(known, pid)
				}
			}

			for pid := range known {
				list.Delete(pid)
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return list, nil
}

func KillProcess(pid int32) error {
	process, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("error getting process: %w", err)
	}

	return process.Kill()
}
