# Instance Selection Algorithm Benchmarking & Simulation

## Overview

This document describes how to test, simulate, and benchmark the Azure instance selection and bin-packing logic, and how to demonstrate the benefits of the new selection algorithm compared to a naive or legacy approach.

## How to Test

### 1. Unit and Simulation Tests

- Run all unit and simulation tests to verify correctness and basic behavior:
  ```bash
  go test ./pkg/resolver/ -v
  ```

- These tests include:
  - Single workload selection (general, CPU, memory, IO-optimized)
  - Bin-packing of multiple workloads onto VMs

### 2. Simulating Benefits

- To demonstrate the benefits of the new selection logic, run the bin-packing simulation with a set of synthetic or real workloads.
- Compare the following metrics:
  - **Total number of VMs used**
  - **Total cost (sum of PricePerHour for all VMs used)**
  - **Resource utilization (CPU/memory usage per VM)**
- Compare results for:
  - Naive selection (e.g., always pick the largest or smallest VM)
  - New selection algorithm (with filtering, scoring, and bin-packing)

- Example: Add a test or main function that loads a workload trace, runs both algorithms, and prints a summary table.

### 3. Output Example

```
Strategy: GeneralPurpose
Total VMs used: 5
Total cost: $1.20/hr
Average CPU utilization: 85%
Average Memory utilization: 80%

Strategy: Naive (smallest VM)
Total VMs used: 9
Total cost: $1.80/hr
Average CPU utilization: 45%
Average Memory utilization: 40%
```

## Using Public Cloud Traces for Benchmarking

To make the simulation realistic, use real-world workload traces:

### 1. Google Cluster Data (GCD)
- Contains millions of jobs with resource requirements and constraints.
- [Google Cluster Data](https://github.com/google/cluster-data)
- Use: Parse task resource requests and simulate bin-packing.

### 2. Azure Public Dataset
- Microsoft's VM workload traces from Azure.
- [Azure VM Workload Traces](https://github.com/Azure/AzurePublicDataset)
- Use: Parse VM deployment requests (vCPU, memory, duration) and simulate scheduling.

### 3. Alibaba Cluster Trace
- Resource utilization patterns from Alibaba's production clusters.
- [Alibaba Cluster Trace](https://github.com/alibaba/clusterdata)
- Use: Parse job resource requirements and simulate bin-packing.

## How to Run a Benchmark

1. **Download and preprocess a trace dataset** (e.g., CSV or JSON).
2. **Write a Go program or test** that:
   - Loads the trace into a slice of `WorkloadProfile`.
   - Defines a set of available `AzureInstanceSpec` (real or synthetic).
   - Runs `BinPackWorkloads` with both the new and naive selection logic.
   - Outputs a summary table with metrics (VMs used, cost, utilization).
3. **Compare results** to show the improvement in efficiency and cost.

## Example: Adding a Benchmark Test

You can add a test like this in `pkg/resolver/instance_types_benchmark_test.go`:

```go
func TestBenchmark_BinPackingWithTrace(t *testing.T) {
    // Load workloads from a trace file (CSV/JSON parsing not shown)
    workloads := LoadWorkloadsFromTrace("azure_trace.csv")
    candidates := LoadAzureInstanceSpecs("azure_skus.json")
    result := BinPackWorkloads(workloads, candidates, StrategyGeneralPurpose)
    naiveResult := BinPackWorkloadsNaive(workloads, candidates)
    fmt.Printf("New algorithm: VMs=%d, Cost=%.2f\n", len(result.VMs), TotalCost(result.VMs))
    fmt.Printf("Naive: VMs=%d, Cost=%.2f\n", len(naiveResult.VMs), TotalCost(naiveResult.VMs))
}
```

## Next Steps

- Implement trace file parsers for public datasets.
- Add benchmark tests as shown above.
- Document and visualize results to demonstrate the benefits of the new selection logic.
