package resolver

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// TraceSource represents a public trace dataset.
type TraceSource string

const (
	TraceGoogle   TraceSource = "google"
	TraceAzure    TraceSource = "azure"
	TraceAlibaba  TraceSource = "alibaba"
)

/*
DownloadTrace downloads and caches a trace file from a public dataset.
If the file is a .gz, but the download is not actually gzipped (e.g. due to proxy or error), it will
detect and fix the file extension to avoid gzip: invalid header errors.
*/
func DownloadTrace(source TraceSource, destDir string) (string, error) {
	var url, filename string
	switch source {
	case TraceGoogle:
		url = "https://storage.googleapis.com/clusterdata-2019-2/clusterdata-2019-2-task-events.csv.gz"
		filename = "google_clusterdata_2019.csv.gz"
	case TraceAzure:
		url = "https://azureopendatastorage.blob.core.windows.net/azurepublicdataset/azure_vm_workload.csv"
		filename = "azure_vm_workload.csv"
	case TraceAlibaba:
		url = "https://github.com/alibaba/clusterdata/raw/master/cluster-trace-micro-2018.csv"
		filename = "alibaba_cluster_trace_2018.csv"
	default:
		return "", errors.New("unknown trace source")
	}
	destPath := filepath.Join(destDir, filename)
	if _, err := os.Stat(destPath); err == nil {
		// Check if .gz file is actually not gzipped (fix for invalid header)
		if strings.HasSuffix(destPath, ".gz") {
			isGz, err := isGzipFile(destPath)
			if err == nil && !isGz {
				// Rename to .csv and use that
				newPath := strings.TrimSuffix(destPath, ".gz") + ".csv"
				os.Rename(destPath, newPath)
				return newPath, nil
			}
		}
		return destPath, nil // already downloaded and valid
	}
	fmt.Printf("Downloading %s to %s...\n", url, destPath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	// Check if .gz file is actually not gzipped (fix for invalid header)
	if strings.HasSuffix(destPath, ".gz") {
		isGz, err := isGzipFile(destPath)
		if err == nil && !isGz {
			newPath := strings.TrimSuffix(destPath, ".gz") + ".csv"
			os.Rename(destPath, newPath)
			return newPath, nil
		}
	}
	return destPath, nil
}

// isGzipFile checks if a file is a valid gzip file by reading its header.
func isGzipFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	var buf [2]byte
	_, err = f.Read(buf[:])
	if err != nil {
		return false, err
	}
	// Gzip files start with 0x1f 0x8b
	return buf[0] == 0x1f && buf[1] == 0x8b, nil
}

/*
LoadWorkloadsFromTrace parses a trace file into a slice of WorkloadProfile.
Supports Google, Azure, and Alibaba public traces (robust parsing).
Handles .gz files for Google trace.
*/
func LoadWorkloadsFromTrace(tracePath string, source TraceSource, maxRows int) ([]WorkloadProfile, error) {
	var r io.Reader
	f, err := os.Open(tracePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r = f

	// Handle .gz for Google trace
	if source == TraceGoogle && strings.HasSuffix(tracePath, ".gz") {
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		r = gzr
	}

	workloads := make([]WorkloadProfile, 0, maxRows)
	csvr := csv.NewReader(r)
	header, err := csvr.Read()
	if err != nil {
		return nil, err
	}

	switch source {
	case TraceGoogle:
		// Google trace: columns: ... requested_cpu, requested_memory, ...
		// Find column indices
		cpuIdx, memIdx := -1, -1
		for i, col := range header {
			if col == "requested_cpu" {
				cpuIdx = i
			}
			if col == "requested_memory" {
				memIdx = i
			}
		}
		if cpuIdx == -1 || memIdx == -1 {
			return nil, errors.New("could not find requested_cpu/requested_memory columns")
		}
		for i := 0; i < maxRows; i++ {
			row, err := csvr.Read()
			if err != nil {
				break
			}
			cpu, _ := strconv.ParseFloat(row[cpuIdx], 64)
			mem, _ := strconv.ParseFloat(row[memIdx], 64)
			if cpu == 0 && mem == 0 {
				continue
			}
			workloads = append(workloads, WorkloadProfile{
				CPURequirements:    int(cpu / 1000), // convert to cores
				MemoryRequirements: mem / 1024,      // convert to GiB
			})
		}
	case TraceAzure:
		// Azure trace: columns: vCPUs, memoryGB, ...
		cpuIdx, memIdx := -1, -1
		for i, col := range header {
			if strings.Contains(strings.ToLower(col), "vcpu") {
				cpuIdx = i
			}
			if strings.Contains(strings.ToLower(col), "memory") {
				memIdx = i
			}
		}
		if cpuIdx == -1 || memIdx == -1 {
			return nil, errors.New("could not find vCPU/memory columns")
		}
		for i := 0; i < maxRows; i++ {
			row, err := csvr.Read()
			if err != nil {
				break
			}
			cpu, _ := strconv.Atoi(row[cpuIdx])
			mem, _ := strconv.ParseFloat(row[memIdx], 64)
			if cpu == 0 && mem == 0 {
				continue
			}
			workloads = append(workloads, WorkloadProfile{
				CPURequirements:    cpu,
				MemoryRequirements: mem,
			})
		}
	case TraceAlibaba:
		// Alibaba trace: columns: ... cpu, mem, ...
		cpuIdx, memIdx := -1, -1
		for i, col := range header {
			if strings.ToLower(col) == "cpu" {
				cpuIdx = i
			}
			if strings.ToLower(col) == "mem" {
				memIdx = i
			}
		}
		if cpuIdx == -1 || memIdx == -1 {
			return nil, errors.New("could not find cpu/mem columns")
		}
		for i := 0; i < maxRows; i++ {
			row, err := csvr.Read()
			if err != nil {
				break
			}
			cpu, _ := strconv.Atoi(row[cpuIdx])
			mem, _ := strconv.ParseFloat(row[memIdx], 64)
			if cpu == 0 && mem == 0 {
				continue
			}
			workloads = append(workloads, WorkloadProfile{
				CPURequirements:    cpu,
				MemoryRequirements: mem,
			})
		}
	default:
		return nil, errors.New("unknown trace source")
	}
	return workloads, nil
}

// LoadAzureInstanceSpecs loads Azure VM SKUs from a JSON file.
func LoadAzureInstanceSpecs(jsonPath string) ([]AzureInstanceSpec, error) {
	data, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	var specs []AzureInstanceSpec
	if err := json.Unmarshal(data, &specs); err != nil {
		return nil, err
	}
	return specs, nil
}

// BinPackWorkloadsNaive is a naive bin-packing: assign each workload to the smallest VM that fits.
func BinPackWorkloadsNaive(workloads WorkloadSet, candidates []AzureInstanceSpec) PackingResult {
	var result PackingResult
	for _, w := range workloads {
		// Find the smallest VM that fits
		var best AzureInstanceSpec
		bestFound := false
		for _, vm := range candidates {
			if vm.VCpus >= w.CPURequirements && vm.MemoryGiB >= w.MemoryRequirements {
				if !bestFound || (vm.VCpus < best.VCpus || (vm.VCpus == best.VCpus && vm.MemoryGiB < best.MemoryGiB)) {
					best = vm
					bestFound = true
				}
			}
		}
		if bestFound {
			result.VMs = append(result.VMs, PackedVM{
				InstanceType: best,
				Workloads:    []WorkloadProfile{w},
			})
		}
	}
	return result
}

// TotalCost computes the total cost per hour for a packing result.
func TotalCost(vms []PackedVM) float64 {
	var sum float64
	for _, vm := range vms {
		sum += vm.InstanceType.PricePerHour
	}
	return sum
}

// AverageUtilization computes average CPU and memory utilization for a packing result.
func AverageUtilization(vms []PackedVM) (cpuUtil, memUtil float64) {
	var totalCPU, usedCPU float64
	var totalMem, usedMem float64
	for _, vm := range vms {
		totalCPU += float64(vm.InstanceType.VCpus)
		totalMem += vm.InstanceType.MemoryGiB
		for _, w := range vm.Workloads {
			usedCPU += float64(w.CPURequirements)
			usedMem += w.MemoryRequirements
		}
	}
	if totalCPU > 0 {
		cpuUtil = usedCPU / totalCPU * 100
	}
	if totalMem > 0 {
		memUtil = usedMem / totalMem * 100
	}
	return
}

type SimulationResult struct {
	VMsUsed   int
	TotalCost float64
	AvgCPU    float64
	AvgMem    float64
}

// RunTraceSimulation downloads, caches, preprocesses, and runs a simulation for a given trace.
func RunTraceSimulation(trace TraceSource, skuPath string, maxRows int) error {
	result, naive, err := RunTraceSimulationWithResults(trace, skuPath, maxRows)
	if err != nil {
		return err
	}
	fmt.Printf("Results:\n")
	fmt.Printf("New algorithm: VMs=%d, Cost=%.2f/hr\n", result.VMsUsed, result.TotalCost)
	fmt.Printf("  Avg CPU utilization: %.1f%%, Avg Mem utilization: %.1f%%\n", result.AvgCPU, result.AvgMem)
	fmt.Printf("Naive: VMs=%d, Cost=%.2f/hr\n", naive.VMsUsed, naive.TotalCost)
	fmt.Printf("  Avg CPU utilization: %.1f%%, Avg Mem utilization: %.1f%%\n", naive.AvgCPU, naive.AvgMem)
	return nil
}

/*
RunTraceSimulationWithResults returns results for both new and naive algorithms for export/visualization.
If trace == "custom", this function will return an error (use RunCustomWorkloadSimulation).
*/
func RunTraceSimulationWithResults(trace TraceSource, skuPath string, maxRows int) (SimulationResult, SimulationResult, error) {
	if trace == "custom" {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("custom trace not supported here, use RunCustomWorkloadSimulation")
	}
	cacheDir := ".trace_cache"
	os.MkdirAll(cacheDir, 0755)
	tracePath, err := DownloadTrace(trace, cacheDir)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("download trace: %w", err)
	}
	fmt.Printf("Parsing workloads from %s...\n", tracePath)
	workloads, err := LoadWorkloadsFromTrace(tracePath, trace, maxRows)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("parse trace: %w", err)
	}
	fmt.Printf("Loading Azure instance specs from %s...\n", skuPath)
	skus, err := LoadAzureInstanceSpecs(skuPath)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("load skus: %w", err)
	}
	fmt.Printf("Simulating bin-packing with new algorithm...\n")
	result := BinPackWorkloads(workloads, skus, StrategyGeneralPurpose)
	fmt.Printf("Simulating bin-packing with naive algorithm...\n")
	naive := BinPackWorkloadsNaive(workloads, skus)
	cpuU, memU := AverageUtilization(result.VMs)
	cpuU2, memU2 := AverageUtilization(naive.VMs)
	return SimulationResult{
			VMsUsed:   len(result.VMs),
			TotalCost: TotalCost(result.VMs),
			AvgCPU:    cpuU,
			AvgMem:    memU,
		}, SimulationResult{
			VMsUsed:   len(naive.VMs),
			TotalCost: TotalCost(naive.VMs),
			AvgCPU:    cpuU2,
			AvgMem:    memU2,
		}, nil
}

/*
RunCustomWorkloadSimulation loads a custom workload JSON file and runs the simulation.
The JSON file should be an array of objects with CPURequirements and MemoryRequirements.
*/
func RunCustomWorkloadSimulation(workloadsFile string, skuPath string) (SimulationResult, SimulationResult, error) {
	data, err := ioutil.ReadFile(workloadsFile)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("read workloads: %w", err)
	}
	var workloads []WorkloadProfile
	if err := json.Unmarshal(data, &workloads); err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("parse workloads: %w", err)
	}
	fmt.Printf("Loaded %d custom workloads from %s\n", len(workloads), workloadsFile)
	fmt.Printf("Loading Azure instance specs from %s...\n", skuPath)
	skus, err := LoadAzureInstanceSpecs(skuPath)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("load skus: %w", err)
	}
	fmt.Printf("Simulating bin-packing with new algorithm...\n")
	result := BinPackWorkloads(workloads, skus, StrategyGeneralPurpose)
	fmt.Printf("Simulating bin-packing with naive algorithm...\n")
	naive := BinPackWorkloadsNaive(workloads, skus)
	cpuU, memU := AverageUtilization(result.VMs)
	cpuU2, memU2 := AverageUtilization(naive.VMs)
	return SimulationResult{
			VMsUsed:   len(result.VMs),
			TotalCost: TotalCost(result.VMs),
			AvgCPU:    cpuU,
			AvgMem:    memU,
		}, SimulationResult{
			VMsUsed:   len(naive.VMs),
			TotalCost: TotalCost(naive.VMs),
			AvgCPU:    cpuU2,
			AvgMem:    memU2,
		}, nil
}
