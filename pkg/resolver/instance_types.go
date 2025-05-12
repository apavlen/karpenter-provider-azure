package resolver

import (
	"strings"
	"fmt"
)

/*
AzureInstanceSpec describes an Azure VM type and its capabilities.

Instance Selection Algorithm: Input and Output

Input:
- The main input to the instance selection algorithm is a list of candidate Azure VM instance types (`[]AzureInstanceSpec`) and a workload profile (`WorkloadProfile`).
  - `AzureInstanceSpec` describes the properties and capabilities of each VM type (CPU, memory, GPU, zones, features, etc).
  - `WorkloadProfile` describes the requirements of the workload to be scheduled (CPU, memory, GPU, zone, and other constraints).

Output:
- The output is the "best" instance type (`AzureInstanceSpec`) from the candidates that satisfies the workload's requirements and optimizes for cost, fit, and other strategy-specific criteria.
- If no suitable instance is found, the output is an empty `AzureInstanceSpec` (with Name == "").

How it works:
- The algorithm filters the candidate instances to only those that meet the workload's constraints (zone, GPU, features, etc).
- It then scores the filtered instances using a strategy-specific scoring function (e.g., general, CPU, memory, IO intensive).
- The instance with the highest score is selected as the output.

Comparison to AWS Karpenter Instance Selection Logic:

- This repo's instance selection logic is conceptually similar to AWS Karpenter's:
  - Both filter instance types based on workload requirements (zone, GPU, ephemeral disk, etc).
  - Both use a scoring/ranking function to select the "best" instance from the filtered set.
  - Both support pluggable strategies (general, CPU, memory, IO intensive).
  - Both support bin-packing for multi-workload scheduling.

- Differences:
  - AWS Karpenter's implementation is more mature, with more advanced scoring, weighting, and support for constraints like interruption rates, launch templates, and capacity type (spot/on-demand).
  - AWS Karpenter uses a more sophisticated sorting (with sort.Slice and stable sort), while this repo uses a simple selection sort for demonstration.
  - This repo's scoring and filtering logic is extensible but currently simpler and more Azure-specific (e.g., Trusted Launch, Accelerated Networking).
  - AWS Karpenter integrates with AWS APIs for real-time instance availability, pricing, and capacity; this repo would need Azure-specific integrations for parity.

- Summary:
  - The high-level approach (filter, score, select) is the same.
  - This repo is a good starting point and is structurally similar, but would need further enhancements for full feature parity with AWS Karpenter.

Azure-specific requirements and constraints to consider:
- Trusted Launch (TTs): Azure supports Trusted Launch for enhanced security (TPM, vTPM, Secure Boot).
- Accelerated Networking: Some workloads require this for high network throughput/low latency.
- MaxPods: Some VM SKUs have a maximum number of pods they support.
- UltraSSDEnabled: Some VMs support Ultra SSD disks.
- Proximity Placement Groups: For low-latency requirements.
- Regional Quotas: vCPU quotas per family/region.
- Spot Eviction Policy: Spot VMs have different eviction policies.
- Confidential Computing: Some VMs support confidential workloads.
- Ephemeral OS Disk: Some VMs support ephemeral OS disks for faster boot.
- Availability Zones: Not all SKUs are available in all zones.
- GPU/FPGA: Some workloads require specific GPU/FPGA types.

These can be modeled as additional fields and filter functions.
*/

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
	TrustedLaunch          bool // TTs: Trusted Launch support
	AcceleratedNetworking  bool
	MaxPods                int
	UltraSSDEnabled        bool
	ProximityPlacement     bool
	// Add more fields as needed for filtering (e.g., AcceleratedNetworking, MaxPods, etc.)
}

/*
WorkloadProfile describes the requirements for a workload (pod).

Capabilities map can be used for Azure-specific requirements, e.g.:
- TrustedLaunch: "true"
- AcceleratedNetworking: "true"
- MaxPods: "30"
- UltraSSDEnabled: "true"
- ProximityPlacement: "true"
*/
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
	Capabilities       map[string]string // Azure-specific requirements
	// Add more fields as needed for filtering (e.g., labels, taints, etc.)
}

// WorkloadSet represents a set of workloads (pods) to be scheduled.
type WorkloadSet []WorkloadProfile

// PackingResult represents the result of bin-packing: which workloads are assigned to which VMs.
type PackingResult struct {
	VMs []PackedVM
}

type PackedVM struct {
	InstanceType AzureInstanceSpec
	Workloads    []WorkloadProfile
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

// FilterFunc is a function that filters instance types based on requirements.
type FilterFunc func(AzureInstanceSpec, WorkloadProfile) bool

// ScoreFunc is a function that scores instance types for a workload.
type ScoreFunc func(AzureInstanceSpec, WorkloadProfile) float64

// FilterInstanceTypes filters a list of instance types based on a set of filter functions.
func FilterInstanceTypes(candidates []AzureInstanceSpec, workload WorkloadProfile, filters ...FilterFunc) []AzureInstanceSpec {
	var filtered []AzureInstanceSpec
	for _, inst := range candidates {
		ok := true
		for _, filter := range filters {
			if !filter(inst, workload) {
				ok = false
				break
			}
		}
		if ok {
			filtered = append(filtered, inst)
		}
	}
	return filtered
}

// Example filter functions (can be extended)
func FilterByZone(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	if workload.Zone == "" {
		return true
	}
	for _, z := range inst.AvailabilityZones {
		if z == workload.Zone {
			return true
		}
	}
	return false
}

func FilterByGPU(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	if workload.GPURequirements == 0 {
		return true
	}
	if inst.GPUCount < workload.GPURequirements {
		return false
	}
	if workload.GPUType != "" && !strings.EqualFold(inst.GPUType, workload.GPUType) {
		return false
	}
	return true
}

func FilterByEphemeralOS(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	if !workload.RequireEphemeralOS {
		return true
	}
	return inst.EphemeralOSDisk
}

func FilterByTrustedLaunch(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	// If workload requires Trusted Launch, only allow VMs that support it
	if val, ok := workload.Capabilities["TrustedLaunch"]; ok && val == "true" {
		return inst.TrustedLaunch
	}
	return true
}

func FilterByAcceleratedNetworking(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	if val, ok := workload.Capabilities["AcceleratedNetworking"]; ok && val == "true" {
		return inst.AcceleratedNetworking
	}
	return true
}

func FilterByMaxPods(inst AzureInstanceSpec, workload WorkloadProfile) bool {
	if val, ok := workload.Capabilities["MaxPods"]; ok {
		// Parse value as int
		var req int
		_, err := fmt.Sscanf(val, "%d", &req)
		if err == nil && inst.MaxPods > 0 {
			return inst.MaxPods >= req
		}
	}
	return true
}

// Add more filters as needed (e.g., spot, confidential, family, etc.)

// RankInstanceTypes sorts instance types by score (descending).
func RankInstanceTypes(candidates []AzureInstanceSpec, workload WorkloadProfile, score ScoreFunc) []AzureInstanceSpec {
	// Simple selection sort for demonstration; replace with sort.Slice for production.
	out := make([]AzureInstanceSpec, len(candidates))
	copy(out, candidates)
	for i := 0; i < len(out); i++ {
		best := i
		for j := i + 1; j < len(out); j++ {
			if score(out[j], workload) > score(out[best], workload) {
				best = j
			}
		}
		out[i], out[best] = out[best], out[i]
	}
	return out
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

/*
selectWithStrategy is a helper to select the best instance with a given strategy.
This now uses filtering and ranking, similar to AWS Karpenter.
*/
func selectWithStrategy(candidates []AzureInstanceSpec, workload WorkloadProfile, strategy SelectionStrategy) (AzureInstanceSpec, float64) {
	// Compose filters (add more as needed)
	filters := []FilterFunc{
		FilterByZone,
		FilterByGPU,
		FilterByEphemeralOS,
		FilterByTrustedLaunch,
		FilterByAcceleratedNetworking,
		FilterByMaxPods,
		// Add more filters here
	}
	filtered := FilterInstanceTypes(candidates, workload, filters...)

	// Choose scoring function based on strategy
	scoreFunc := func(vm AzureInstanceSpec, w WorkloadProfile) float64 {
		return ScoreInstance(vm, w, strategy)
	}
	ranked := RankInstanceTypes(filtered, workload, scoreFunc)
	if len(ranked) == 0 {
		return AzureInstanceSpec{}, -1
	}
	best := ranked[0]
	return best, scoreFunc(best, workload)
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

// --- Bin-packing (multi-workload scheduling) ---

// BinPackWorkloads assigns workloads to VMs using a first-fit decreasing bin-packing algorithm.
// Returns a PackingResult with the list of VMs and their assigned workloads.
func BinPackWorkloads(workloads WorkloadSet, candidates []AzureInstanceSpec, strategy SelectionStrategy) PackingResult {
	// Sort workloads by descending CPU+Memory demand (naive, can be improved)
	sorted := make(WorkloadSet, len(workloads))
	copy(sorted, workloads)
	// Simple bubble sort for demonstration
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CPURequirements+int(sorted[j].MemoryRequirements) > sorted[i].CPURequirements+int(sorted[i].MemoryRequirements) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	var result PackingResult
	unpacked := make([]bool, len(sorted))

	for {
		// Find the next workload not yet packed
		nextIdx := -1
		for i, packed := range unpacked {
			if !packed {
				nextIdx = i
				break
			}
		}
		if nextIdx == -1 {
			break // all packed
		}
		// For this workload, select the best instance type
		workload := sorted[nextIdx]
		bestVM, _ := selectWithStrategy(candidates, workload, strategy)
		if bestVM.Name == "" {
			break // no suitable VM found
		}
		// Try to pack as many workloads as possible onto this VM
		var packed []WorkloadProfile
		remainingCPU := bestVM.VCpus
		remainingMem := bestVM.MemoryGiB
		for i, w := range sorted {
			if unpacked[i] {
				continue
			}
			if w.CPURequirements <= remainingCPU && w.MemoryRequirements <= remainingMem {
				packed = append(packed, w)
				remainingCPU -= w.CPURequirements
				remainingMem -= w.MemoryRequirements
				unpacked[i] = true
			}
		}
		result.VMs = append(result.VMs, PackedVM{
			InstanceType: bestVM,
			Workloads:    packed,
		})
	}
	return result
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
