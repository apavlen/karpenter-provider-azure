package resolver_test

import (
    "testing"
    "karpenter-provider-azure/pkg/resolver"
)





func TestComputeFit(t *testing.T) {
    vm := resolver.VMType{VCpuCapacity: 8, MemoryCapacity: 32}
    workload := resolver.WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
    fit := resolver.ComputeFit(vm, workload)
    if fit < 0.99 || fit > 1.0 {
        t.Errorf("Expected fit ~1.0, got %v", fit)
    }
}

func TestScoreVM(t *testing.T) {
    vm := resolver.VMType{
        PricePerVCpu:   0.1,
        PricePerGiB:    0.2,
        VCpuCapacity:   8,
        MemoryCapacity: 32,
    }
    workload := resolver.WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
    score := resolver.ScoreVM(vm, workload)
    if score <= 0 {
        t.Errorf("Expected positive score, got %v", score)
    }
}

func TestSelectBestVM(t *testing.T) {
    candidates := []resolver.VMType{
        resolver.VMType{PricePerVCpu: 0.1, PricePerGiB: 0.2, VCpuCapacity: 8, MemoryCapacity: 32},
        resolver.VMType{PricePerVCpu: 0.2, PricePerGiB: 0.3, VCpuCapacity: 4, MemoryCapacity: 16},
    }
    workload := resolver.WorkloadProfile{CPURequirements: 4, MemoryRequirements: 16}
    best := resolver.SelectBestVM(candidates, workload)
    if best.PricePerVCpu != 0.1 {
        t.Errorf("Expected best candidate with PricePerVCpu 0.1, got %v", best.PricePerVCpu)
    }
}
