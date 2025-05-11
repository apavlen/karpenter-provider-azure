package main

import (
	"fmt"
	"math/rand"
	"time"

	"pkg/resolver"
)

func main() {
	// Example Azure instance types (in real use, load from file or API)
	instanceTypes := []resolver.AzureInstanceSpec{
		{
			Name:                  "Standard_D4s_v3",
			VCpus:                 4,
			MemoryGiB:             16,
			StorageGiB:            64,
			PricePerHour:          0.2,
			Family:                "Dsv3",
			Capabilities:          map[string]string{"AcceleratedNetworking": "true"},
			GPUCount:              0,
			GPUType:               "",
			AvailabilityZones:     []string{"1", "2", "3"},
			EphemeralOSDisk:       true,
			NestedVirtualization:  true,
			SpotSupported:         true,
			ConfidentialComputing: false,
			TrustedLaunch:         true,
			AcceleratedNetworking: true,
			MaxPods:               30,
			UltraSSDEnabled:       false,
			ProximityPlacement:    false,
		},
		{
			Name:                  "Standard_NC6s_v3",
			VCpus:                 6,
			MemoryGiB:             112,
			StorageGiB:            340,
			PricePerHour:          1.2,
			Family:                "NCasv3",
			Capabilities:          map[string]string{"GPU": "NVIDIA"},
			GPUCount:              1,
			GPUType:               "NVIDIA",
			AvailabilityZones:     []string{"1", "2"},
			EphemeralOSDisk:       false,
			NestedVirtualization:  false,
			SpotSupported:         true,
			ConfidentialComputing: false,
			TrustedLaunch:         false,
			AcceleratedNetworking: true,
			MaxPods:               40,
			UltraSSDEnabled:       true,
			ProximityPlacement:    false,
		},
		// Add more instance types as needed
	}

	// Example workloads (in real use, load from file or generate)
	workloads := make([]resolver.WorkloadProfile, 0, 10)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		workloads = append(workloads, resolver.WorkloadProfile{
			CPURequirements:    rand.Intn(3) + 1, // 1-3 vCPU
			MemoryRequirements: float64(rand.Intn(8) + 2), // 2-9 GiB
			IORequirements:     float64(rand.Intn(20)),    // 0-19 GiB
			GPURequirements:    0,
			GPUType:            "",
			Zone:               "",
			RequireEphemeralOS: rand.Intn(2) == 0,
			RequireNestedVirt:  rand.Intn(2) == 0,
			RequireSpot:        rand.Intn(2) == 0,
			RequireConfidential: false,
			Capabilities:       map[string]string{},
		})
	}
	// Add a GPU workload
	workloads = append(workloads, resolver.WorkloadProfile{
		CPURequirements:    4,
		MemoryRequirements: 32,
		IORequirements:     100,
		GPURequirements:    1,
		GPUType:            "NVIDIA",
		Zone:               "1",
		RequireEphemeralOS: false,
		RequireNestedVirt:  false,
		RequireSpot:        false,
		RequireConfidential: false,
		Capabilities:       map[string]string{"AcceleratedNetworking": "true"},
	})

	// Run the simulation
	result := resolver.BinPackWorkloads(workloads, instanceTypes, resolver.StrategyGeneralPurpose)

	// Output results
	fmt.Printf("Simulation Results:\n")
	fmt.Printf("Total VMs used: %d\n", len(result.VMs))
	totalCost := 0.0
	for i, vm := range result.VMs {
		vmCost := vm.InstanceType.PricePerHour
		fmt.Printf("VM #%d: %s (vCPUs: %d, Mem: %.1f GiB, GPU: %d, Price: $%.2f/hr)\n",
			i+1, vm.InstanceType.Name, vm.InstanceType.VCpus, vm.InstanceType.MemoryGiB, vm.InstanceType.GPUCount, vmCost)
		fmt.Printf("  Workloads packed: %d\n", len(vm.Workloads))
		for _, w := range vm.Workloads {
			fmt.Printf("    - CPU: %d, Mem: %.1f GiB, GPU: %d\n", w.CPURequirements, w.MemoryRequirements, w.GPURequirements)
		}
		totalCost += vmCost
	}
	fmt.Printf("Total hourly cost: $%.2f\n", totalCost)
}
