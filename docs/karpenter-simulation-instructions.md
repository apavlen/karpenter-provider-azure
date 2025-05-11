# Karpenter Simulation Instructions

## Prerequisites

- Go 1.20+ installed
- This repository cloned locally

## Running the Simulation

1. Build and run the simulation program:

   ```bash
   go run ./cmd/karpenter-sim/main.go
   ```

   (If the simulation driver is in a different location, adjust the path accordingly.)

2. The program will output:
   - The list of VMs selected and their assigned workloads
   - Total number of VMs used
   - Total cost (if implemented)
   - Packing efficiency metrics

## Customizing the Simulation

- Edit the workload and instance type definitions in `cmd/karpenter-sim/main.go` to try different scenarios.
- You can add more test cases or constraints as needed.

## Troubleshooting

- If you encounter build errors, ensure all dependencies are installed and your Go version is up to date.
- For questions, see the design doc: [docs/karpenter-simulation-design.md](karpenter-simulation-design.md)
