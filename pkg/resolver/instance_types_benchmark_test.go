package resolver

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

/*
Benchmarking with Realistic Traces:

This file supports benchmarking the instance selection algorithm using real Azure VM traces
(preprocessed into pod-like workload profiles).

To generate the workload data:
    1. Download and preprocess Azure traces using the provided Python scripts.
    2. Use the output JSON file as input to the Go benchmark.

The benchmark measures:
    - Cost efficiency (estimated monthly cost)
    - Resource utilization (CPU/memory usage ratio)
    - Instance diversity (types of VMs selected)
    - Selection time (algorithm performance)
    - A/B comparison between strategies

To run the benchmark with real Azure trace data:

1. Make sure you have preprocessed Azure trace data in the file `workloads_preprocessed.json` in your project root (or adjust the path in BenchmarkInstanceSelectionWithRealWorkloads).
2. Run the following command to execute the benchmark using the real Azure workloads:

    go test -bench BenchmarkInstanceSelectionWithRealWorkloads -benchmem ./pkg/resolver

This will run only the benchmark that uses the Azure trace data, not the synthetic one.

To visualize results, see the scripts/visualize_benchmark_results.py script.
*/

// WorkloadProfileJSON is a struct for loading preprocessed workloads from JSON.
type WorkloadProfileJSON struct {
	CPURequest      int               `json:"cpu_request"`
	MemoryRequestGi float64           `json:"memory_request_gib"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
}

// loadAzureWorkloads loads preprocessed workload profiles from a JSON file.
func loadAzureWorkloads(path string) []WorkloadProfile {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("failed to open workload file: %v", err))
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var raw []WorkloadProfileJSON
	if err := dec.Decode(&raw); err != nil {
		panic(fmt.Sprintf("failed to decode workload file: %v", err))
	}
	workloads := make([]WorkloadProfile, 0, len(raw))
	for _, w := range raw {
		workloads = append(workloads, WorkloadProfile{
			CPURequirements:    w.CPURequest,
			MemoryRequirements: w.MemoryRequestGi,
			Capabilities:       map[string]string{"AcceleratedNetworking": "true"},
		})
	}
	return workloads
}

// BenchmarkInstanceSelectionWithRealWorkloads runs the benchmark using real Azure workloads.
func BenchmarkInstanceSelectionWithRealWorkloads(b *testing.B) {
	// You may need to adjust the path to your preprocessed data file
	workloadFile := "workloads_preprocessed.json"
	if _, err := os.Stat(workloadFile); os.IsNotExist(err) {
		b.Fatalf("ERROR: BenchmarkInstanceSelectionWithRealWorkloads requires %s but it was not found.\n"+
			"Please generate this file using the preprocessing script before running the benchmark.\n"+
			"See scripts/preprocess_azure_traces.py for details.", workloadFile)
		return
	}
	workloads := loadAzureWorkloads(workloadFile)

	// Generate synthetic instance types (in production, load from Azure API or static list)
	numInstances := 100
	candidates := make([]AzureInstanceSpec, numInstances)
	for i := 0; i < numInstances; i++ {
		candidates[i] = randomInstanceSpec(i)
	}

	strategies := map[string]InstanceSelector{
		"General":      &GeneralPurposeSelector{},
		"CPUOptimized": &CPUStrategySelector{},
		"MemOptimized": &MemoryStrategySelector{},
	}

	b.ResetTimer()
	for name, selector := range strategies {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			selectionCounts := make(map[string]int)
			for i := 0; i < b.N; i++ {
				w := workloads[i%len(workloads)]
				best, _ := selector.Select(candidates, w)
				selectionCounts[best.Name]++
			}
			// Output summary after benchmark
			b.Logf("Strategy: %s, Unique VM types selected: %d", name, len(selectionCounts))
			for vm, count := range selectionCounts {
				b.Logf("  VM: %s, selected %d times", vm, count)
			}
		})
	}
}

// --- Existing synthetic benchmark for reference ---

func randomInstanceSpec(i int) AzureInstanceSpec {
	return AzureInstanceSpec{
		Name:                  fmt.Sprintf("Standard_D%d_v4", i),
		VCpus:                 rand.Intn(64) + 2,
		MemoryGiB:             float64(rand.Intn(256) + 4),
		StorageGiB:            float64(rand.Intn(2000) + 32),
		PricePerHour:          rand.Float64()*10 + 0.05,
		Family:                "Dsv4",
		Capabilities:          map[string]string{"AcceleratedNetworking": "true"},
		GPUCount:              rand.Intn(2),
		GPUType:               "",
		AvailabilityZones:     []string{"1", "2", "3"},
		EphemeralOSDisk:       rand.Intn(2) == 0,
		NestedVirtualization:  rand.Intn(2) == 0,
		SpotSupported:         rand.Intn(2) == 0,
		ConfidentialComputing: rand.Intn(2) == 0,
		TrustedLaunch:         rand.Intn(2) == 0,
		AcceleratedNetworking: rand.Intn(2) == 0,
		MaxPods:               rand.Intn(250) + 30,
		UltraSSDEnabled:       rand.Intn(2) == 0,
		ProximityPlacement:    rand.Intn(2) == 0,
	}
}

func randomWorkloadProfile() WorkloadProfile {
	return WorkloadProfile{
		CPURequirements:     rand.Intn(16) + 1,
		MemoryRequirements:  float64(rand.Intn(64) + 1),
		IORequirements:      float64(rand.Intn(100) + 1),
		GPURequirements:     rand.Intn(2),
		GPUType:             "",
		Zone:                fmt.Sprintf("%d", rand.Intn(3)+1),
		RequireEphemeralOS:  rand.Intn(2) == 0,
		RequireNestedVirt:   rand.Intn(2) == 0,
		RequireSpot:         rand.Intn(2) == 0,
		RequireConfidential: rand.Intn(2) == 0,
		Capabilities:        map[string]string{"AcceleratedNetworking": "true"},
	}
}

func BenchmarkInstanceSelection(b *testing.B) {
	// This benchmark uses only synthetic (randomly generated) traces, not real Azure data.
	// The real Azure trace benchmark is BenchmarkInstanceSelectionWithRealWorkloads above.
	rand.Seed(time.Now().UnixNano())
	numInstances := 1000
	numWorkloads := 1000

	candidates := make([]AzureInstanceSpec, numInstances)
	for i := 0; i < numInstances; i++ {
		candidates[i] = randomInstanceSpec(i)
	}

	workloads := make([]WorkloadProfile, numWorkloads)
	for i := 0; i < numWorkloads; i++ {
		workloads[i] = randomWorkloadProfile()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := workloads[i%numWorkloads]
		_ = SelectBestInstance(candidates, w)
	}
}
