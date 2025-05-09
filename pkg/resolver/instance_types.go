package resolver

import "strings"

// AzureInstanceSpec describes an Azure VM type and its capabilities.
type AzureInstanceSpec struct {
	Name                   string
	VCpus                  int
	MemoryGiB              float64
	StorageGiB             float64
	PricePerHour           float64
	Family                 string
	Capabilities           map[string]string
	GPUCount               int
	GPUType                string
	AvailabilityZones      []string
	EphemeralOSDisk        bool
	NestedVirtualization   bool
	SpotSupported          bool
	ConfidentialComputing  bool
}

// WorkloadProfile describes the requirements for a workload.
type WorkloadProfile struct {
	CPURequirements    int
	MemoryRequirements float64
	IORequirements     float64 // optional, can be 0
	GPURequirements    int     // optional, can be 0
	GPUType            string  // optional, can be ""
	Zone               string  // optional, can be ""
	RequireEphemeralOS bool
	RequireNestedVirt  bool
	RequireSpot        bool
	RequireConfidential bool
}

// SelectionStrategy defines the type of selection algorithm.
type SelectionStrategy string

const (
	StrategyGeneralPurpose SelectionStrategy = "general"
	StrategyCPUIntensive   SelectionStrategy = "cpu"
	StrategyMemoryIntensive SelectionStrategy = "memory"
	StrategyIOIntensive    SelectionStrategy = "io"
)

/*
InstanceSelector is the interface for pluggable selection algorithms.
*/
type InstanceSelector interface {
	Select(candidates []AzureInstanceSpec, workload WorkloadProfile) (AzureInstanceSpec, float64)
}

// GeneralPurposeSelector implements InstanceSelector for general workloads.
type GeneralPurposeSelector struct{}

func (s *GeneralPurposeSelector) Select(candidates []AzureInstanceSpec, workload WorkloadProfile) (AzureInstanceSpec, float64) {
	return selectWithStrategy(candidates, workload, StrategyGeneralPurpose)
}

// CPUStrategySelector implements InstanceSelector for CPU-optimized workloads.
type CPUStrategySelector struct{}

func (s *CPUStrategySelector) Select(candidates []AzureInstanceSpec, workload WorkloadProfile) (AzureInstanceSpec, float64) {
	return selectWithStrategy(candidates, workload, StrategyCPUIntensive)
}

// MemoryStrategySelector implements InstanceSelector for memory-optimized workloads.
type MemoryStrategySelector struct{}

func (s *MemoryStrategySelector) Select(candidates []AzureInstanceSpec, workload WorkloadProfile) (AzureInstanceSpec, float64) {
	return selectWithStrategy(candidates, workload, StrategyMemoryIntensive)
}

// IOStrategySelector implements InstanceSelector for IO-optimized workloads.
type IOStrategySelector struct{}

func (s *IOStrategySelector) Select(candidates []AzureInstanceSpec, workload WorkloadProfile) (AzureInstanceSpec, float64) {
	return selectWithStrategy(candidates, workload, StrategyIOIntensive)
}

// selectWithStrategy is a helper to select the best instance with a given strategy.
func selectWithStrategy(candidates []AzureInstanceSpec, workload WorkloadProfile, strategy SelectionStrategy) (AzureInstanceSpec, float64) {
	var best AzureInstanceSpec
	bestScore := -1.0
	for _, candidate := range candidates {
		score := ScoreInstance(candidate, workload, strategy)
		if score > bestScore {
			bestScore = score
			best = candidate
		}
	}
	return best, bestScore
}

// ScoreInstance scores a VM for a workload and strategy.
func ScoreInstance(vm AzureInstanceSpec, workload WorkloadProfile, strategy SelectionStrategy) float64 {
	// Cost efficiency: lower is better
	costEfficiency := 1.0 / (vm.PricePerHour + 0.01)
	resourceFit := ComputeFit(vm, workload)
	availabilityScore := zoneScore(vm, workload.Zone)
	gpuScore := gpuFit(vm, workload)
	ephemeralScore := boolScore(vm.EphemeralOSDisk, workload.RequireEphemeralOS)
	nestedVirtScore := boolScore(vm.NestedVirtualization, workload.RequireNestedVirt)
	spotScore := boolScore(vm.SpotSupported, workload.RequireSpot)
	confidentialScore := boolScore(vm.ConfidentialComputing, workload.RequireConfidential)

	// Strategy-specific weighting
	switch strategy {
	case StrategyCPUIntensive:
		return 0.5*cpuFit(vm, workload) + 0.2*costEfficiency + 0.1*resourceFit + 0.1*availabilityScore + 0.1*gpuScore
	case StrategyMemoryIntensive:
		return 0.5*memFit(vm, workload) + 0.2*costEfficiency + 0.1*resourceFit + 0.1*availabilityScore + 0.1*gpuScore
	case StrategyIOIntensive:
		return 0.5*ioFit(vm, workload) + 0.2*costEfficiency + 0.1*resourceFit + 0.1*availabilityScore + 0.1*gpuScore
	default:
		// General purpose: balance all
		return 0.3*costEfficiency + 0.2*resourceFit + 0.1*availabilityScore + 0.1*gpuScore +
			0.1*ephemeralScore + 0.1*nestedVirtScore + 0.05*spotScore + 0.05*confidentialScore
	}
}

// ComputeFit returns a value in [0,1] for how well the VM fits the workload.
func ComputeFit(vm AzureInstanceSpec, workload WorkloadProfile) float64 {
	cpu := cpuFit(vm, workload)
	mem := memFit(vm, workload)
	io := ioFit(vm, workload)
	// Use the lowest fit as the limiting factor
	fit := cpu
	if mem < fit {
		fit = mem
	}
	if io < fit {
		fit = io
	}
	if fit > 1.0 {
		fit = 1.0
	}
	return fit
}

func cpuFit(vm AzureInstanceSpec, workload WorkloadProfile) float64 {
	if workload.CPURequirements == 0 {
		return 1.0
	}
	return min(float64(vm.VCpus)/float64(workload.CPURequirements), 1.0)
}

func memFit(vm AzureInstanceSpec, workload WorkloadProfile) float64 {
	if workload.MemoryRequirements == 0 {
		return 1.0
	}
	return min(vm.MemoryGiB/workload.MemoryRequirements, 1.0)
}

func ioFit(vm AzureInstanceSpec, workload WorkloadProfile) float64 {
	if workload.IORequirements == 0 {
		return 1.0
	}
	return min(vm.StorageGiB/workload.IORequirements, 1.0)
}

func gpuFit(vm AzureInstanceSpec, workload WorkloadProfile) float64 {
	if workload.GPURequirements == 0 {
		return 1.0
	}
	if vm.GPUCount < workload.GPURequirements {
		return 0.0
	}
	if workload.GPUType != "" && !strings.EqualFold(vm.GPUType, workload.GPUType) {
		return 0.0
	}
	return 1.0
}

func zoneScore(vm AzureInstanceSpec, zone string) float64 {
	if zone == "" {
		return 1.0
	}
	for _, z := range vm.AvailabilityZones {
		if z == zone {
			return 1.0
		}
	}
	return 0.0
}

func boolScore(vmHas, required bool) float64 {
	if !required {
		return 1.0
	}
	if vmHas {
		return 1.0
	}
	return 0.0
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

/*
SelectBestInstance is a convenience function for general-purpose selection.
*/
func SelectBestInstance(candidates []AzureInstanceSpec, workload WorkloadProfile) AzureInstanceSpec {
	selector := &GeneralPurposeSelector{}
	best, _ := selector.Select(candidates, workload)
	return best
}

// SelectBestInstanceWithStrategy allows selection with a specific strategy.
func SelectBestInstanceWithStrategy(candidates []AzureInstanceSpec, workload WorkloadProfile, strategy SelectionStrategy) AzureInstanceSpec {
	var selector InstanceSelector
	switch strategy {
	case StrategyCPUIntensive:
		selector = &CPUStrategySelector{}
	case StrategyMemoryIntensive:
		selector = &MemoryStrategySelector{}
	case StrategyIOIntensive:
		selector = &IOStrategySelector{}
	default:
		selector = &GeneralPurposeSelector{}
	}
	best, _ := selector.Select(candidates, workload)
	return best
}
