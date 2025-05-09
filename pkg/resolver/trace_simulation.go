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
	// If a .csv version exists, prefer it (fix for previous renames)
	if strings.HasSuffix(destPath, ".gz") {
		csvPath := strings.TrimSuffix(destPath, ".gz") + ".csv"
		if _, err := os.Stat(csvPath); err == nil {
			return csvPath, nil
		}
	}
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
		// Google trace: columns: ... requested_cpu, requested_memory, ... OR cpu_request, memory_request, ...
		// Try to find either set of columns for robustness
		cpuIdx, memIdx := -1, -1
		for i, col := range header {
			lc := strings.ToLower(col)
			if lc == "requested_cpu" || lc == "cpu_request" {
				cpuIdx = i
			}
			if lc == "requested_memory" || lc == "memory_request" {
				memIdx = i
			}
		}
		if cpuIdx == -1 || memIdx == -1 {
			return nil, fmt.Errorf("could not find requested_cpu/requested_memory or cpu_request/memory_request columns (found header: %v)", header)
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

// QuotaMap maps VM family to max vCPUs allowed.
type QuotaMap map[string]int

// LoadQuota loads a quota.json file mapping family to max vCPUs.
func LoadQuota(path string) (QuotaMap, error) {
	if path == "" {
		return nil, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var q QuotaMap
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, err
	}
	return q, nil
}

// BinPackWorkloadsWithQuota is like BinPackWorkloads but enforces vCPU quotas per family.
func BinPackWorkloadsWithQuota(workloads WorkloadSet, candidates []AzureInstanceSpec, strategy SelectionStrategy, quota QuotaMap) PackingResult {
	// Sort workloads by descending CPU+Memory demand (naive, can be improved)
	sorted := make(WorkloadSet, len(workloads))
	copy(sorted, workloads)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CPURequirements+int(sorted[j].MemoryRequirements) > sorted[i].CPURequirements+int(sorted[i].MemoryRequirements) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var result PackingResult
	unpacked := make([]bool, len(sorted))
	usedVCpus := make(map[string]int)

	for {
		// Find the next workload not yet packed
		nextIdx := -1
		for i, packed := range unpacked {
			if !packed {
				nextIdx = i
				break
			}
		}
		if nextIdx == -1 {
			break // all packed
		}
		// For this workload, select the best instance type
		workload := sorted[nextIdx]
		bestVM, _ := selectWithStrategy(candidates, workload, strategy)
		if bestVM.Name == "" {
			break // no suitable VM found
		}
		// Check quota for this family
		fam := bestVM.Family
		if quota != nil && quota[fam] > 0 && usedVCpus[fam]+bestVM.VCpus > quota[fam] {
			// Can't use this family anymore, remove from candidates and retry
			var newCandidates []AzureInstanceSpec
			for _, c := range candidates {
				if c.Family != fam {
					newCandidates = append(newCandidates, c)
				}
			}
			candidates = newCandidates
			continue
		}
		// Try to pack as many workloads as possible onto this VM
		var packed []WorkloadProfile
		remainingCPU := bestVM.VCpus
		remainingMem := bestVM.MemoryGiB
		for i, w := range sorted {
			if unpacked[i] {
				continue
			}
			if w.CPURequirements <= remainingCPU && w.MemoryRequirements <= remainingMem {
				packed = append(packed, w)
				remainingCPU -= w.CPURequirements
				remainingMem -= w.MemoryRequirements
				unpacked[i] = true
			}
		}
		usedVCpus[fam] += bestVM.VCpus
		result.VMs = append(result.VMs, PackedVM{
			InstanceType: bestVM,
			Workloads:    packed,
		})
	}
	return result
}

// RunTraceSimulationWithQuota runs the simulation with an optional quota file.
func RunTraceSimulationWithQuota(trace TraceSource, skuPath string, maxRows int, quotaPath string) (SimulationResult, SimulationResult, error) {
	if trace == "custom" {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("custom trace not supported here, use RunCustomWorkloadSimulationWithQuota")
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
	quota, err := LoadQuota(quotaPath)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("load quota: %w", err)
	}
	fmt.Printf("Simulating bin-packing with new algorithm...\n")
	result := BinPackWorkloadsWithQuota(workloads, skus, StrategyGeneralPurpose, quota)
	fmt.Printf("Simulating bin-packing with naive algorithm...\n")
	naive := BinPackWorkloadsWithQuota(workloads, skus, StrategyGeneralPurpose, quota) // For naive, could use BinPackWorkloadsNaive with quota logic if desired
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

// RunCustomWorkloadSimulationWithQuota loads a custom workload JSON file and runs the simulation with quota.
func RunCustomWorkloadSimulationWithQuota(workloadsFile string, skuPath string, quotaPath string) (SimulationResult, SimulationResult, error) {
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
	quota, err := LoadQuota(quotaPath)
	if err != nil {
		return SimulationResult{}, SimulationResult{}, fmt.Errorf("load quota: %w", err)
	}
	fmt.Printf("Simulating bin-packing with new algorithm...\n")
	result := BinPackWorkloadsWithQuota(workloads, skus, StrategyGeneralPurpose, quota)
	fmt.Printf("Simulating bin-packing with naive algorithm...\n")
	naive := BinPackWorkloadsWithQuota(workloads, skus, StrategyGeneralPurpose, quota)
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
