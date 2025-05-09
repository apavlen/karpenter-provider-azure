package resolver

import (
	"testing"
)

func TestGeneralPurposeSelector_Simple(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_E4_v4",
			VCpus:        4,
			MemoryGiB:    32,
			StorageGiB:   100,
			PricePerHour: 0.20,
			Family:       "Standard_E",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2", "3"},
		},
		{
			Name:         "Standard_NC6",
			VCpus:        6,
			MemoryGiB:    56,
			StorageGiB:   380,
			PricePerHour: 0.90,
			Family:       "Standard_NC",
			Capabilities: map[string]string{},
			GPUCount:     1,
			GPUType:      "NVIDIA",
			AvailabilityZones: []string{"2"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    2,
		MemoryRequirements: 7,
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_D2_v4" {
		t.Errorf("expected Standard_D2_v4, got %s", best.Name)
	}
}

func TestGeneralPurposeSelector_GPU(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_NC6",
			VCpus:        6,
			MemoryGiB:    56,
			StorageGiB:   380,
			PricePerHour: 0.90,
			Family:       "Standard_NC",
			Capabilities: map[string]string{},
			GPUCount:     1,
			GPUType:      "NVIDIA",
			AvailabilityZones: []string{"2"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    4,
		MemoryRequirements: 16,
		GPURequirements:    1,
		GPUType:            "NVIDIA",
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_NC6" {
		t.Errorf("expected Standard_NC6, got %s", best.Name)
	}
}

func TestGeneralPurposeSelector_Zone(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_E4_v4",
			VCpus:        4,
			MemoryGiB:    32,
			StorageGiB:   100,
			PricePerHour: 0.20,
			Family:       "Standard_E",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"3"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    2,
		MemoryRequirements: 7,
		Zone:               "3",
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_E4_v4" {
		t.Errorf("expected Standard_E4_v4, got %s", best.Name)
	}
}
package resolver

import (
	"testing"
)

func TestGeneralPurposeSelector_Simple(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_E4_v4",
			VCpus:        4,
			MemoryGiB:    32,
			StorageGiB:   100,
			PricePerHour: 0.20,
			Family:       "Standard_E",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2", "3"},
		},
		{
			Name:         "Standard_NC6",
			VCpus:        6,
			MemoryGiB:    56,
			StorageGiB:   380,
			PricePerHour: 0.90,
			Family:       "Standard_NC",
			Capabilities: map[string]string{},
			GPUCount:     1,
			GPUType:      "NVIDIA",
			AvailabilityZones: []string{"2"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    2,
		MemoryRequirements: 7,
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_D2_v4" {
		t.Errorf("expected Standard_D2_v4, got %s", best.Name)
	}
}

func TestGeneralPurposeSelector_GPU(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_NC6",
			VCpus:        6,
			MemoryGiB:    56,
			StorageGiB:   380,
			PricePerHour: 0.90,
			Family:       "Standard_NC",
			Capabilities: map[string]string{},
			GPUCount:     1,
			GPUType:      "NVIDIA",
			AvailabilityZones: []string{"2"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    4,
		MemoryRequirements: 16,
		GPURequirements:    1,
		GPUType:            "NVIDIA",
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_NC6" {
		t.Errorf("expected Standard_NC6, got %s", best.Name)
	}
}

func TestGeneralPurposeSelector_Zone(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{
			Name:         "Standard_D2_v4",
			VCpus:        2,
			MemoryGiB:    8,
			StorageGiB:   50,
			PricePerHour: 0.10,
			Family:       "Standard_D",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"1", "2"},
		},
		{
			Name:         "Standard_E4_v4",
			VCpus:        4,
			MemoryGiB:    32,
			StorageGiB:   100,
			PricePerHour: 0.20,
			Family:       "Standard_E",
			Capabilities: map[string]string{},
			AvailabilityZones: []string{"3"},
		},
	}

	workload := WorkloadProfile{
		CPURequirements:    2,
		MemoryRequirements: 7,
		Zone:               "3",
	}

	best := SelectBestInstance(candidates, workload)
	if best.Name != "Standard_E4_v4" {
		t.Errorf("expected Standard_E4_v4, got %s", best.Name)
	}
}

// New: Simulate CPU-optimized and Memory-optimized selection
func TestCPUStrategySelector(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{Name: "cpu1", VCpus: 16, MemoryGiB: 16, PricePerHour: 0.4},
		{Name: "mem1", VCpus: 4, MemoryGiB: 32, PricePerHour: 0.4},
	}
	workload := WorkloadProfile{CPURequirements: 8, MemoryRequirements: 8}
	best := SelectBestInstanceWithStrategy(candidates, workload, StrategyCPUIntensive)
	if best.Name != "cpu1" {
		t.Errorf("expected cpu1 for CPU-optimized, got %s", best.Name)
	}
}

func TestMemoryStrategySelector(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{Name: "cpu1", VCpus: 16, MemoryGiB: 16, PricePerHour: 0.4},
		{Name: "mem1", VCpus: 4, MemoryGiB: 32, PricePerHour: 0.4},
	}
	workload := WorkloadProfile{CPURequirements: 2, MemoryRequirements: 24}
	best := SelectBestInstanceWithStrategy(candidates, workload, StrategyMemoryIntensive)
	if best.Name != "mem1" {
		t.Errorf("expected mem1 for Memory-optimized, got %s", best.Name)
	}
}
