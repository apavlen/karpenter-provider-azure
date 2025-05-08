package resolver

type VMType struct {
    PricePerVCpu   float64
    PricePerGiB    float64
    VCpuCapacity   int
    MemoryCapacity int
}

type WorkloadProfile struct {
    CPURequirements    int
    MemoryRequirements int
}

func ScoreVM(vm VMType, workload WorkloadProfile) float64 {
    costEfficiency := 1.0 / (vm.PricePerVCpu + vm.PricePerGiB)
    resourceFit := ComputeFit(vm, workload) // We'll define this next
    availabilityScore := 0.9 // placeholder for now
    return 0.4*costEfficiency + 0.3*resourceFit + 0.3*availabilityScore
}

func ComputeFit(vm VMType, workload WorkloadProfile) float64 {
    // Compute a simple measure of resource fit based on CPU and Memory ratios.
    cpuFit := float64(vm.VCpuCapacity) / float64(workload.CPURequirements)
    memFit := float64(vm.MemoryCapacity) / float64(workload.MemoryRequirements)
    // Use the lower ratio as the fitness measure.
    fit := cpuFit
    if memFit < fit {
        fit = memFit
    }
    if fit > 1.0 {
        fit = 1.0
    }
    return fit
}

func SelectBestVM(candidates []VMType, workload WorkloadProfile) VMType {
    var best VMType
    bestScore := -1.0
    for _, candidate := range candidates {
        score := ScoreVM(candidate, workload)
        if score > bestScore {
            bestScore = score
            best = candidate
        }
    }
    return best
}
