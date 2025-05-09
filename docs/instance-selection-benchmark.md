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

1. **Download and preprocess a trace dataset** (e.g., CSV or JSON) using the provided simulation tool.
2. **Run the simulation** using the provided Go function to compare the new and naive selection logic.
3. **Review the output summary table** with metrics (VMs used, cost, utilization) to see the improvement in efficiency and cost.

## Example: Running the Simulation

You can run the simulation from a Go main or test file like this:

```go
import "github.com/Azure/karpenter-provider-azure/pkg/resolver"

func main() {
    // Download, cache, preprocess, and simulate with Google trace and a local Azure SKU file
    err := resolver.RunTraceSimulation(resolver.TraceGoogle, "azure_skus.json", 1000)
    if err != nil {
        panic(err)
    }
}
```

The output will look like:

```
Downloading https://storage.googleapis.com/clusterdata-2019-2/clusterdata-2019-2-task-events.csv.gz to .trace_cache/google_clusterdata_2019.csv.gz...
Parsing workloads from .trace_cache/google_clusterdata_2019.csv.gz...
Loading Azure instance specs from azure_skus.json...
Simulating bin-packing with new algorithm...
Simulating bin-packing with naive algorithm...
Results:
New algorithm: VMs=5, Cost=1.20/hr
  Avg CPU utilization: 85.0%, Avg Mem utilization: 80.0%
Naive: VMs=9, Cost=1.80/hr
  Avg CPU utilization: 45.0%, Avg Mem utilization: 40.0%
```

## Next Steps

- Use the provided `RunTraceSimulation` function to benchmark with different public datasets (Google, Azure, Alibaba).
- Add or update your Azure SKU JSON file to match your region or requirements.
- Document and visualize results to demonstrate the benefits of the new selection logic.

## Fetching Azure SKU Data

To fetch and preprocess Azure VM SKU data for simulation, use the provided script:

```bash
python3 scripts/fetch_azure_skus.py > azure_skus.json
```

This will create a `azure_skus.json` file suitable for use with the simulation.

## Running the Simulation CLI

To run the simulation with a real trace and your SKU file:

```bash
go run ./cmd/instance-selection-sim/ -trace google -sku azure_skus.json -max 1000
```

You can also use `-trace azure` or `-trace alibaba` for other datasets.

## Visualizing and Exporting Results

To further analyze and visualize the results, you can export the simulation output to a CSV file for plotting or reporting.  
The CLI supports an optional `-out results.csv` flag to write a summary of the simulation results.

Example:

```bash
go run ./cmd/instance-selection-sim/ -trace google -sku azure_skus.json -max 1000 -out results.csv
```

The CSV will contain:

| Strategy      | VMs Used | Total Cost | Avg CPU Util (%) | Avg Mem Util (%) |
|---------------|----------|------------|------------------|------------------|
| NewAlgorithm  | 5        | 1.20       | 85.0             | 80.0             |
| Naive         | 9        | 1.80       | 45.0             | 40.0             |

You can then use tools like Excel, Google Sheets, or Python/pandas/matplotlib to visualize the efficiency gains.
