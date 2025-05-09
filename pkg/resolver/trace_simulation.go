package resolver

import (
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
)

// TraceSource represents a public trace dataset.
type TraceSource string

const (
	TraceGoogle   TraceSource = "google"
	TraceAzure    TraceSource = "azure"
	TraceAlibaba  TraceSource = "alibaba"
)

// DownloadTrace downloads and caches a trace file from a public dataset.
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
		return destPath, nil // already downloaded
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
	return destPath, nil
}

// LoadWorkloadsFromTrace parses a trace file into a slice of WorkloadProfile.
// Supports Google, Azure, and Alibaba public traces (basic parsing).
func LoadWorkloadsFromTrace(tracePath string, source TraceSource, maxRows int) ([]WorkloadProfile, error) {
	f, err := os.Open(tracePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var workloads []WorkloadProfile
	switch source {
	case TraceGoogle:
		// Google trace: CSV, columns: ... requested_cpu, requested_memory, ...
		r := csv.NewReader(f)
		_, _ = r.Read() // skip header
		for i := 0; i < maxRows; i++ {
			row, err := r.Read()
			if err != nil {
				break
			}
			// Example: CPU in millicores, memory in MB
			cpu, _ := strconv.ParseFloat(row[9], 64)
			mem, _ := strconv.ParseFloat(row[10], 64)
			if cpu == 0 && mem == 0 {
				continue
			}
			workloads = append(workloads, WorkloadProfile{
				CPURequirements:    int(cpu / 1000), // convert to cores
				MemoryRequirements: mem / 1024,      // convert to GiB
			})
		}
	case TraceAzure:
		// Azure trace: CSV, columns: vCPUs, memoryGB, ...
		r := csv.NewReader(f)
		_, _ = r.Read() // skip header
		for i := 0; i < maxRows; i++ {
			row, err := r.Read()
			if err != nil {
				break
			}
			cpu, _ := strconv.Atoi(row[0])
			mem, _ := strconv.ParseFloat(row[1], 64)
			if cpu == 0 && mem == 0 {
				continue
			}
			workloads = append(workloads, WorkloadProfile{
				CPURequirements:    cpu,
				MemoryRequirements: mem,
			})
		}
	case TraceAlibaba:
		// Alibaba trace: CSV, columns: ... cpu, mem, ...
		r := csv.NewReader(f)
		_, _ = r.Read() // skip header
		for i := 0; i < maxRows; i++ {
			row, err := r.Read()
			if err != nil {
				break
			}
			cpu, _ := strconv.Atoi(row[2])
			mem, _ := strconv.ParseFloat(row[3], 64)
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

// RunTraceSimulation downloads, caches, preprocesses, and runs a simulation for a given trace.
func RunTraceSimulation(trace TraceSource, skuPath string, maxRows int) error {
	cacheDir := ".trace_cache"
	os.MkdirAll(cacheDir, 0755)
	tracePath, err := DownloadTrace(trace, cacheDir)
	if err != nil {
		return fmt.Errorf("download trace: %w", err)
	}
	fmt.Printf("Parsing workloads from %s...\n", tracePath)
	workloads, err := LoadWorkloadsFromTrace(tracePath, trace, maxRows)
	if err != nil {
		return fmt.Errorf("parse trace: %w", err)
	}
	fmt.Printf("Loading Azure instance specs from %s...\n", skuPath)
	skus, err := LoadAzureInstanceSpecs(skuPath)
	if err != nil {
		return fmt.Errorf("load skus: %w", err)
	}
	fmt.Printf("Simulating bin-packing with new algorithm...\n")
	result := BinPackWorkloads(workloads, skus, StrategyGeneralPurpose)
	fmt.Printf("Simulating bin-packing with naive algorithm...\n")
	naive := BinPackWorkloadsNaive(workloads, skus)
	fmt.Printf("Results:\n")
	fmt.Printf("New algorithm: VMs=%d, Cost=%.2f/hr\n", len(result.VMs), TotalCost(result.VMs))
	cpuU, memU := AverageUtilization(result.VMs)
	fmt.Printf("  Avg CPU utilization: %.1f%%, Avg Mem utilization: %.1f%%\n", cpuU, memU)
	fmt.Printf("Naive: VMs=%d, Cost=%.2f/hr\n", len(naive.VMs), TotalCost(naive.VMs))
	cpuU, memU = AverageUtilization(naive.VMs)
	fmt.Printf("  Avg CPU utilization: %.1f%%, Avg Mem utilization: %.1f%%\n", cpuU, memU)
	return nil
}
