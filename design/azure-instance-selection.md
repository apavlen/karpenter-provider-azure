# Azure Instance Selection Algorithm Design

## Overview

This document describes the design for an extensible Azure instance type selection algorithm for Karpenter on Azure, inspired by the logic in [karpenter-provider-aws](https://github.com/aws/karpenter-provider-aws/tree/main/pkg/providers). The goal is to select the best VM instance type for a given workload, based on CPU, memory, and other requirements, with the ability to extend to more advanced algorithms (e.g., prediction, IO, spot, etc.) in the future.

## Goals

- **Flexible instance selection**: Support for CPU-optimized, memory-optimized, IO-optimized, and general-purpose selection strategies.
- **Extensibility**: Easy to add new selection algorithms (e.g., ML-based, cost-aware, spot-aware).
- **Simulation and benchmarking**: Ability to simulate scheduling decisions and benchmark with real or synthetic traces.
- **Testability**: Comprehensive unit and integration tests to validate selection logic and demonstrate benefits.

## Key Concepts

- **AzureInstanceSpec**: Struct describing Azure VM instance types (CPU, memory, storage, price, etc.).
- **WorkloadProfile**: Struct describing workload requirements (CPU, memory, IO, etc.).
- **InstanceSelector**: Interface for pluggable selection algorithms.
- **SelectionStrategy**: Enum or type for different strategies (CPU, memory, IO, prediction, etc.).
- **Simulation**: Tooling to simulate scheduling decisions and compare strategies.

## Data Model

### AzureInstanceSpec

- Name (string)
- VCpus (int)
- MemoryGiB (float64)
- StorageGiB (float64, optional)
- PricePerHour (float64)
- Family (string, e.g., "Standard_D", "Standard_E")
- Capabilities (map[string]string) // e.g., "acceleratedNetworking": "true"

### WorkloadProfile

- CPU (int)
- Memory (float64)
- IO (optional, float64)
- Labels/taints (optional)

## Algorithm

- **General-purpose**: Score based on fit (CPU/memory) and cost.
- **CPU-optimized**: Prefer instance types with high CPU:memory ratio.
- **Memory-optimized**: Prefer instance types with high memory:CPU ratio.
- **IO-optimized**: Prefer instance types with high storage or network throughput.
- **Extensible**: New strategies can be plugged in via the InstanceSelector interface.

## Extensibility

- New algorithms can be added by implementing the InstanceSelector interface.
- Selection strategy can be chosen via config or API.

## Simulation & Testing

- Simulate scheduling with synthetic or real traces.
- Compare strategies by cost, fit, and resource utilization.
- Unit tests for scoring and selection logic.
- Integration tests with sample instance specs and workloads.

## Next Steps

1. Define data structures for AzureInstanceSpec and WorkloadProfile.
2. Implement InstanceSelector interface and basic strategies (CPU, memory, general).
3. Add simulation tooling and tests.
4. Prepare for real trace integration.
