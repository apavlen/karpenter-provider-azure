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

You can run the simulation and get results as follows:

```bash
go run ./cmd/instance-selection-sim/ -trace google -sku azure_skus.json -max 1000 -out results.csv
```

- This will run the simulation using the Google trace and your Azure SKU file.
- The results will be written to `results.csv` in the current directory.

The output will look like:

```
Results:
New algorithm: VMs=5, Cost=1.20/hr
  Avg CPU utilization: 85.0%, Avg Mem utilization: 80.0%
Naive: VMs=9, Cost=1.80/hr
  Avg CPU utilization: 45.0%, Avg Mem utilization: 40.0%
Results written to results.csv
```

You can then plot the results with:

```bash
python3 scripts/plot_simulation_results.py results.csv
```

- The `results.csv` file will contain a summary table with the results for each strategy.
- The plot script will generate bar charts comparing VMs used, total cost, and utilization for each strategy.

The `results.csv` file is your main output artifact for further analysis and visualization.

## Built-in Visualization

A helper script is provided to plot the results directly:

```bash
python3 scripts/plot_simulation_results.py results.csv
```

This will generate bar charts comparing VMs used, total cost, and utilization for each strategy.

---

## Advanced: Regional SKU Fetching, Quota Simulation, and Custom Workload Generation

### 1. Fetching SKUs for Different Azure Regions

To benchmark with SKUs from a specific Azure region, edit the `API` variable in `scripts/fetch_azure_skus.py`:

```python
API = "https://prices.azure.com/api/retail/prices?$filter=serviceName eq 'Virtual Machines' and armRegionName eq 'westeurope'"
```

Then run:

```bash
python3 scripts/fetch_azure_skus.py > azure_skus_westeurope.json
```

### 2. Simulating Quota Constraints

To simulate quota constraints (e.g., max vCPUs per family/region), you can:
- Add a `quota.json` file with limits per VM family, e.g.:
  ```json
  {
    "Standard_D": 32,
    "Standard_E": 64
  }
  ```
- Extend the Go simulation to read this file and filter out SKUs or limit the number of VMs per family.

#### Example: Using Quota Constraints

1. Create a `quota.json` file as above.
2. Run the simulation with the new `-quota quota.json` flag:
   ```bash
   go run ./cmd/instance-selection-sim/ -trace google -sku azure_skus.json -max 1000 -quota quota.json
   ```
3. The simulation will ensure that the total vCPUs used per family does not exceed the quota.

### 3. Custom Workload Generation

To generate synthetic workloads for stress-testing:

```python
import random, json

def gen_workloads(n):
    return [
        {
            "CPURequirements": random.choice([1,2,4,8]),
            "MemoryRequirements": random.choice([2,4,8,16,32])
        }
        for _ in range(n)
    ]

with open("synthetic_workloads.json", "w") as f:
    json.dump(gen_workloads(1000), f, indent=2)
```

Then, add a loader in Go to read this JSON and run the simulation.

### 4. Example: Running with Custom Workloads

```bash
go run ./cmd/instance-selection-sim/ -trace custom -sku azure_skus.json -max 1000 -workloads synthetic_workloads.json
```

---

## Future Work

- Add support for quota-aware scheduling and reporting.
- Add support for spot/eviction simulation.
- Add support for compliance/region/family constraints.
- Add more advanced bin-packing and prediction strategies.
- Integrate with real Azure API for live SKU/pricing updates.
