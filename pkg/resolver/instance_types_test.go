package resolver_test

import (
	"testing"

	. "github.com/Azure/karpenter-provider-azure/pkg/resolver"
)

func TestComputeFit(t *testing.T) {
	vm := AzureInstanceSpec{VCpus: 8, MemoryGiB: 32}
	workload := WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
	fit := ComputeFit(vm, workload)
	if fit < 0.99 || fit > 1.0 {
		t.Errorf("Expected fit ~1.0, got %v", fit)
	}
}

func TestScoreInstance(t *testing.T) {
	vm := AzureInstanceSpec{
		Name:        "Standard_D4_v4",
		VCpus:       8,
		MemoryGiB:   32,
		PricePerHour: 0.2,
	}
	workload := WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
	score := ScoreInstance(vm, workload, StrategyGeneralPurpose)
	if score <= 0 {
		t.Errorf("Expected positive score, got %v", score)
	}
}

func TestSelectBestInstance(t *testing.T) {
	candidates := []AzureInstanceSpec{
		{Name: "A", VCpus: 8, MemoryGiB: 32, PricePerHour: 0.2},
		{Name: "B", VCpus: 4, MemoryGiB: 16, PricePerHour: 0.1},
	}
	workload := WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
	best := SelectBestInstance(candidates, workload)
	if best.Name != "B" {
		t.Errorf("Expected best candidate with Name B, got %v", best.Name)
	}
}
