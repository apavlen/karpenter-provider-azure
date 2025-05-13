package resolver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"fmt"
)

/*
NOTE: If you see "undefined: WorkloadProfile" or similar errors,
make sure that instance_types.go is in the same package ("resolver")
and that your build/test command includes both files, e.g.:

    go test -bench=Benchmark -benchmem ./pkg/resolver

Running `go test` on a single file (e.g., ..._test.go) will not include
other non-test files in the package. Always run tests at the package level.
*/

// WorkloadJSON is the struct for loading workloads_preprocessed.json
type WorkloadJSON struct {
	Name              string             `json:"name"`
	CPURequest        int                `json:"cpu_request"`
	MemoryRequestGiB  float64            `json:"memory_request_gib"`
	CPUUsage          float64            `json:"cpu_usage"`
	MemUsage          float64            `json:"mem_usage"`
	StartTime         string             `json:"start_time"`
	EndTime           string             `json:"end_time"`
	Labels            map[string]string  `json:"labels"`
	Annotations       map[string]string  `json:"annotations"`
}

/*
Helper to load workloads_preprocessed.json and convert to []WorkloadProfile.
If limit > 0, returns at most limit workloads. If limit == 0, returns all.
*/
func loadWorkloadsFromJSONWithLimit(path string, limit int) ([]WorkloadProfile, error) {
	// Try the provided path first
	f, err := os.Open(path)
	if err != nil {
		// If not found, try looking in parent directory and in testdata/
		altPaths := []string{
			filepath.Join("..", path),
			filepath.Join("testdata", path),
			filepath.Join("..", "testdata", path),
		}
		for _, alt := range altPaths {
			f, err = os.Open(alt)
			if err == nil {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()
	var raw []WorkloadJSON
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}
	var out []WorkloadProfile
	count := 0
	for _, w := range raw {
		if limit > 0 && count >= limit {
			break
		}
		out = append(out, WorkloadProfile{
			CPURequirements:    w.CPURequest,
			MemoryRequirements: w.MemoryRequestGiB,
			// Optionally, you could use CPUUsage/MemUsage for more advanced benchmarking
			Capabilities: map[string]string{
				"workload_type": w.Labels["workload_type"],
			},
		})
		count++
	}
	return out, nil
}

// Backward-compatible: original function loads all workloads
func loadWorkloadsFromJSON(path string) ([]WorkloadProfile, error) {
	return loadWorkloadsFromJSONWithLimit(path, 0)
}

func dummyInstanceTypes() []AzureInstanceSpec {
	return []AzureInstanceSpec{
		{Name: "Standard_D2_v3", VCpus: 2, MemoryGiB: 8, PricePerHour: 0.1, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_D4_v3", VCpus: 4, MemoryGiB: 16, PricePerHour: 0.2, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_D8_v3", VCpus: 8, MemoryGiB: 32, PricePerHour: 0.4, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_D16_v3", VCpus: 16, MemoryGiB: 64, PricePerHour: 0.8, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_D32_v3", VCpus: 32, MemoryGiB: 128, PricePerHour: 1.6, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_E4s_v3", VCpus: 4, MemoryGiB: 32, PricePerHour: 0.25, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_E8s_v3", VCpus: 8, MemoryGiB: 64, PricePerHour: 0.5, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_E16s_v3", VCpus: 16, MemoryGiB: 128, PricePerHour: 1.0, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_F4s_v2", VCpus: 4, MemoryGiB: 8, PricePerHour: 0.22, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_F8s_v2", VCpus: 8, MemoryGiB: 16, PricePerHour: 0.44, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_F16s_v2", VCpus: 16, MemoryGiB: 32, PricePerHour: 0.88, AvailabilityZones: []string{"1", "2", "3"}},
		{Name: "Standard_NC6", VCpus: 6, MemoryGiB: 56, GPUCount: 1, GPUType: "K80", PricePerHour: 0.9, AvailabilityZones: []string{"1"}},
		{Name: "Standard_NC12", VCpus: 12, MemoryGiB: 112, GPUCount: 2, GPUType: "K80", PricePerHour: 1.8, AvailabilityZones: []string{"1"}},
		{Name: "Standard_NC24", VCpus: 24, MemoryGiB: 224, GPUCount: 4, GPUType: "K80", PricePerHour: 3.6, AvailabilityZones: []string{"1"}},
		{Name: "Standard_NV6", VCpus: 6, MemoryGiB: 56, GPUCount: 1, GPUType: "M60", PricePerHour: 1.0, AvailabilityZones: []string{"1"}},
		{Name: "Standard_NV12", VCpus: 12, MemoryGiB: 112, GPUCount: 2, GPUType: "M60", PricePerHour: 2.0, AvailabilityZones: []string{"1"}},
		{Name: "Standard_NV24", VCpus: 24, MemoryGiB: 224, GPUCount: 4, GPUType: "M60", PricePerHour: 4.0, AvailabilityZones: []string{"1"}},
	}
}

// Benchmark instance selection for each workload in the trace
// Optionally limit the number of workloads by setting the WORKLOAD_LIMIT environment variable.
func BenchmarkInstanceSelection_RealTrace(b *testing.B) {
	limit := 0
	if v := os.Getenv("WORKLOAD_LIMIT"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
		if limit > 0 {
			b.Logf("Limiting workloads to %d (via WORKLOAD_LIMIT)", limit)
		}
	}
	workloads, err := loadWorkloadsFromJSONWithLimit("workloads_preprocessed.json", limit)
	if err != nil {
		b.Fatalf("failed to load workloads: %v", err)
	}
	instances := dummyInstanceTypes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range workloads {
			_ = SelectBestInstance(instances, w)
		}
	}
}

type BinPackingAlgorithm func(workloads WorkloadSet, candidates []AzureInstanceSpec, strategy SelectionStrategy) PackingResult

// First-fit decreasing (current default)
func BinPackWorkloadsFFD(workloads WorkloadSet, candidates []AzureInstanceSpec, strategy SelectionStrategy) PackingResult {
	return BinPackWorkloads(workloads, candidates, strategy)
}

// Naive one-workload-per-VM (worst case, for comparison)
func BinPackWorkloadsNaiveAlgo(workloads WorkloadSet, candidates []AzureInstanceSpec, strategy SelectionStrategy) PackingResult {
	var result PackingResult
	for _, w := range workloads {
		bestVM, _ := selectWithStrategy(candidates, w, strategy)
		if bestVM.Name != "" {
			result.VMs = append(result.VMs, PackedVM{
				InstanceType: bestVM,
				Workloads:    []WorkloadProfile{w},
			})
		}
	}
	return result
}

// Benchmark bin-packing for the full trace, comparing algorithms
// Optionally limit the number of workloads by setting the WORKLOAD_LIMIT environment variable.
func BenchmarkBinPacking_RealTrace(b *testing.B) {
	limit := 0
	if v := os.Getenv("WORKLOAD_LIMIT"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
		if limit > 0 {
			b.Logf("Limiting workloads to %d (via WORKLOAD_LIMIT)", limit)
		}
	}
	workloads, err := loadWorkloadsFromJSONWithLimit("workloads_preprocessed.json", limit)
	if err != nil {
		b.Fatalf("failed to load workloads: %v", err)
	}
	instances := dummyInstanceTypes()

	algorithms := []struct {
		name string
		fn   BinPackingAlgorithm
	}{
		{"FirstFitDecreasing", BinPackWorkloadsFFD},
		{"NaiveOnePerVM", BinPackWorkloadsNaiveAlgo},
	}

	for _, algo := range algorithms {
		b.Run(algo.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = algo.fn(workloads, instances, StrategyGeneralPurpose)
			}
		})
	}
}

func TestPackingEfficiencyAndCostReport_RealTrace(t *testing.T) {
	limit := 0
	if v := os.Getenv("WORKLOAD_LIMIT"); v != "" {
		fmt.Sscanf(v, "%d", &limit)
		if limit > 0 {
			t.Logf("Limiting workloads to %d (via WORKLOAD_LIMIT)", limit)
		}
	}
	workloads, err := loadWorkloadsFromJSONWithLimit("workloads_preprocessed.json", limit)
	if err != nil {
		t.Fatalf("failed to load workloads: %v", err)
	}
	instances := dummyInstanceTypes()

	t.Logf("Test running with %d workloads", len(workloads))
	t.Logf("Starting BinPackWorkloads with %d workloads and %d instance types", len(workloads), len(instances))
	result := BinPackWorkloads(workloads, instances, StrategyGeneralPurpose)
	fmt.Printf("Packed %d VMs for %d workloads\n", len(result.VMs), len(workloads))
	totalCPUUsed := 0
	totalMemUsed := 0.0
	totalCPUCap := 0
	totalMemCap := 0.0
	totalCost := 0.0

	fmt.Printf("\n%-20s %-10s %-10s %-10s %-10s %-10s %-10s %-10s\n", "VM Type", "vCPU Used", "vCPU Cap", "Mem Used", "Mem Cap", "CPU Util", "Mem Util", "Cost/hr")
	for _, vm := range result.VMs {
		vmCPU := 0
		vmMem := 0.0
		for _, w := range vm.Workloads {
			vmCPU += w.CPURequirements
			vmMem += w.MemoryRequirements
		}
		totalCPUUsed += vmCPU
		totalMemUsed += vmMem
		totalCPUCap += vm.InstanceType.VCpus
		totalMemCap += vm.InstanceType.MemoryGiB
		totalCost += vm.InstanceType.PricePerHour
		cpuUtil := 100 * float64(vmCPU) / float64(vm.InstanceType.VCpus)
		memUtil := 100 * vmMem / vm.InstanceType.MemoryGiB
		fmt.Printf("%-20s %-10d %-10d %-10.1f %-10.1f %-10.1f %-10.1f $%-9.2f\n",
			vm.InstanceType.Name, vmCPU, vm.InstanceType.VCpus, vmMem, vm.InstanceType.MemoryGiB, cpuUtil, memUtil, vm.InstanceType.PricePerHour)
	}
	fmt.Printf("\nTotal used: %d vCPU / %.1f GiB\n", totalCPUUsed, totalMemUsed)
	fmt.Printf("Total capacity: %d vCPU / %.1f GiB\n", totalCPUCap, totalMemCap)
	if totalCPUCap > 0 {
		fmt.Printf("Overall CPU Utilization: %.1f%%\n", 100*float64(totalCPUUsed)/float64(totalCPUCap))
	} else {
		fmt.Printf("Overall CPU Utilization: N/A (totalCPUCap=0)\n")
	}
	if totalMemCap > 0 {
		fmt.Printf("Overall Memory Utilization: %.1f%%\n", 100*totalMemUsed/totalMemCap)
	} else {
		fmt.Printf("Overall Memory Utilization: N/A (totalMemCap=0)\n")
	}
	fmt.Printf("Total hourly cost: $%.2f\n", totalCost)
	if len(result.VMs) > 0 {
		fmt.Printf("Average cost per VM: $%.2f/hr\n", totalCost/float64(len(result.VMs)))
	}
	if len(workloads) > 0 {
		fmt.Printf("Average cost per workload: $%.4f/hr\n", totalCost/float64(len(workloads)))
	}
	t.Logf("Test completed successfully, packed %d VMs", len(result.VMs))
}
