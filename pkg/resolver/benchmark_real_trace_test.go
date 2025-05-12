package resolver

import (
	"encoding/json"
	"os"
	"testing"
	"fmt"
)

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

// Helper to load workloads_preprocessed.json and convert to []WorkloadProfile
func loadWorkloadsFromJSON(path string) ([]WorkloadProfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var raw []WorkloadJSON
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return nil, err
	}
	var out []WorkloadProfile
	for _, w := range raw {
		out = append(out, WorkloadProfile{
			CPURequirements:    w.CPURequest,
			MemoryRequirements: w.MemoryRequestGiB,
			// Optionally, you could use CPUUsage/MemUsage for more advanced benchmarking
			Capabilities: map[string]string{
				"workload_type": w.Labels["workload_type"],
			},
		})
	}
	return out, nil
}

// Dummy instance types for demonstration; replace with your real instance catalog
func dummyInstanceTypes() []AzureInstanceSpec {
	return []AzureInstanceSpec{
		{
			Name: "Standard_D2_v3", VCpus: 2, MemoryGiB: 8, PricePerHour: 0.1, AvailabilityZones: []string{"1", "2", "3"},
		},
		{
			Name: "Standard_D4_v3", VCpus: 4, MemoryGiB: 16, PricePerHour: 0.2, AvailabilityZones: []string{"1", "2", "3"},
		},
		{
			Name: "Standard_D8_v3", VCpus: 8, MemoryGiB: 32, PricePerHour: 0.4, AvailabilityZones: []string{"1", "2", "3"},
		},
		{
			Name: "Standard_D16_v3", VCpus: 16, MemoryGiB: 64, PricePerHour: 0.8, AvailabilityZones: []string{"1", "2", "3"},
		},
	}
}

// Benchmark instance selection for each workload in the trace
func BenchmarkInstanceSelection_RealTrace(b *testing.B) {
	workloads, err := loadWorkloadsFromJSON("workloads_preprocessed.json")
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

// Benchmark bin-packing for the full trace
func BenchmarkBinPacking_RealTrace(b *testing.B) {
	workloads, err := loadWorkloadsFromJSON("workloads_preprocessed.json")
	if err != nil {
		b.Fatalf("failed to load workloads: %v", err)
	}
	instances := dummyInstanceTypes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BinPackWorkloads(workloads, instances, StrategyGeneralPurpose)
	}
}

// Optionally, a test to print packing results for inspection
func TestPrintBinPackingResult_RealTrace(t *testing.T) {
	workloads, err := loadWorkloadsFromJSON("workloads_preprocessed.json")
	if err != nil {
		t.Fatalf("failed to load workloads: %v", err)
	}
	instances := dummyInstanceTypes()
	result := BinPackWorkloads(workloads, instances, StrategyGeneralPurpose)
	fmt.Printf("Packed %d VMs for %d workloads\n", len(result.VMs), len(workloads))
	for i, vm := range result.VMs {
		fmt.Printf("VM %d: %s, %d workloads\n", i+1, vm.InstanceType.Name, len(vm.Workloads))
	}
}
