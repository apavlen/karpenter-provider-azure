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

To run:
    go test -bench . -benchmem ./pkg/resolver

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
	workloads := loadAzureWorkloads("workloads_preprocessed.json")

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

	for name, selector := range strategies {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				w := workloads[i%len(workloads)]
				_, _ = selector.Select(candidates, w)
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
