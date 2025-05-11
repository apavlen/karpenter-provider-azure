# Karpenter Simulation Design Doc

## Background

We are developing a simulation framework for Karpenter's Azure instance selection and bin-packing logic. The goal is to validate and improve the algorithms that select VM types and pack workloads (pods) onto them, ensuring cost efficiency, resource fit, and support for Azure-specific features.

## What We Are Trying to Do

- **Simulate Karpenter's scheduling and bin-packing logic** for Azure, using realistic instance types and workload profiles.
- **Identify and fix issues** in the selection and packing algorithms, such as:
  - Suboptimal instance selection (e.g., picking too expensive or underpowered VMs)
  - Poor bin-packing efficiency (e.g., too many VMs used)
  - Incorrect handling of Azure-specific constraints (zones, GPUs, Trusted Launch, etc.)
- **Enable rapid iteration and testing** of new selection strategies and filters.

## Design Overview

- **Instance and Workload Modeling:** Use `AzureInstanceSpec` and `WorkloadProfile` structs to represent VM types and workloads.
- **Filtering:** Apply a series of filter functions to eliminate incompatible VM types for each workload.
- **Scoring and Ranking:** Score remaining candidates using strategy-specific functions (general, CPU, memory, IO).
- **Bin-Packing:** Use a first-fit decreasing algorithm to assign workloads to VMs, simulating Karpenter's packing logic.
- **Simulation Driver:** A Go program that loads a set of instance types and workloads, runs the simulation, and outputs results (e.g., number of VMs used, cost, packing efficiency).

## Action Items

1. **Implement a simulation driver** (Go program) that:
   - Loads a set of Azure instance types (can be hardcoded or loaded from a file).
   - Loads a set of workload profiles (can be hardcoded or loaded from a file).
   - Runs the bin-packing simulation using the logic in `pkg/resolver/instance_types.go`.
   - Outputs a summary of results (VMs used, total cost, packing efficiency, etc.).
2. **Add test scenarios** for different workload mixes and constraints (e.g., GPU, zones, spot, etc.).
3. **Validate correctness** by comparing simulation output to expected results.
4. **Iterate on filters and scoring** to improve selection and packing.
5. **Document how to run the simulation** and interpret results.

## Running the Simulation

See [docs/karpenter-simulation-instructions.md](karpenter-simulation-instructions.md) for step-by-step instructions.

## Future Work

- Integrate with real Azure SKU data.
- Add support for more Azure-specific constraints.
- Visualize packing results.
- Automate regression testing for changes to selection logic.
