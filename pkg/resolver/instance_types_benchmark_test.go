package resolver

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

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
