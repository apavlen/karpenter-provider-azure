package resolver

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

/*
Benchmark Results (example run):

goos: linux
goarch: amd64
pkg: github.com/Azure/karpenter-provider-azure/pkg/resolver
cpu: Intel(R) Xeon(R) CPU E5-2673 v4 @ 2.30GHz
BenchmarkInstanceSelection-8   	     140	   9056469 ns/op	  137050 B/op	      10 allocs/op
PASS
ok  	github.com/Azure/karpenter-provider-azure/pkg/resolver	2.137s

With -benchmem:

goos: linux
goarch: amd64
pkg: github.com/Azure/karpenter-provider-azure/pkg/resolver
cpu: Intel(R) Xeon(R) CPU E5-2673 v4 @ 2.30GHz
BenchmarkInstanceSelection-8   	     148	   9357721 ns/op	  153163 B/op	      10 allocs/op
PASS
ok  	github.com/Azure/karpenter-provider-azure/pkg/resolver	2.208s

Benchmarking with Realistic Traces:

For more realistic benchmarking and profiling, you can use traces from:
- Kubernetes scheduler logs (real pod resource requests and scheduling events).
- Public datasets such as the Alibaba Cluster Trace Program (https://github.com/alibaba/clusterdata), Google Borg traces, or Microsoft Azure VM traces.
- Production cluster pod specs (exported as JSON/YAML).

To use a real trace, parse the trace file and convert each workload event into a WorkloadProfile struct.
This allows benchmarking the instance selection logic under realistic, production-like load and diversity.
*/

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
