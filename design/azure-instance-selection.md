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

## Suggestions for Enhancement

### 1. Azure-specific Considerations

- **VM Series Specialization**: Add explicit logic for Azure VM families (e.g., Dv5, Ev5, N-series) to leverage their unique characteristics.
- **Azure Spot VMs**: Incorporate Azure's spot VM implementation, which differs from AWS, including eviction rates and pricing.
- **Azure Quotas**: Make the selection algorithm quota-aware, considering regional vCPU and resource quotas.

### 2. Data Model Enhancements

- **GPU Support**: Add GPU count and type for N-series VMs.
- **Availability Zones**: Track which zones each VM type is available in.
- **Ephemeral OS Disk Support**: Flag for ephemeral OS disk capability.
- **Nested Virtualization Support**: Flag for nested virtualization support.

### 3. Algorithm Refinements

- **Compliance & Regulatory**: Add strategies for compliance (e.g., confidential computing, region restrictions).
- **Multi-dimensional Scoring**: Use a weighted scoring system to balance cost, fit, compliance, and other factors.
- **Fallback Mechanism**: Implement fallback logic when optimal instances are unavailable.

### 4. Implementation Considerations

- **Cache Azure VM SKU Information**: Cache VM specs to avoid repeated Azure API calls.
- **Versioning**: Plan for Azure Compute API versioning.
- **Rate Limiting**: Handle Azure API rate limits gracefully.

### 5. Testing Strategy

- **Regional Variations**: Test with different Azure regions due to VM availability differences.
- **Benchmark Against Actual Pricing**: Validate selection with real pricing data.
- **Failure Scenarios**: Test behavior when preferred or required instances are unavailable.

## Implementation Phases

**Phase 1: Core Infrastructure**
- Basic data structures and interfaces
- Simple scoring algorithm for general-purpose VMs
- Unit test framework

**Phase 2: Azure Integration**
- Azure API client for VM information
- Azure-specific constraints and capabilities
- Integration tests with mocked Azure responses

**Phase 3: Selection Strategies**
- Implement specialized strategies (CPU, memory, IO)
- Cost optimization logic
- Benchmark framework

**Phase 4: Advanced Features**
- Spot VM support
- Predictive scaling
- Production readiness (logging, monitoring, etc.)

## Next Steps

1. Define data structures for AzureInstanceSpec and WorkloadProfile.
2. Implement InstanceSelector interface and basic strategies (CPU, memory, general).
3. Add simulation tooling and tests.
4. Prepare for real trace integration.
# Azure Instance Selection Algorithm Design

> **Azure-specific requirements and constraints considered in this implementation:**
>
> - **Trusted Launch (TTs):** Support for confidential/secure VM boot (TPM, vTPM, Secure Boot).
> - **Accelerated Networking:** High network throughput/low latency.
> - **MaxPods:** Maximum number of pods per VM SKU.
> - **UltraSSDEnabled:** Support for Ultra SSD disks.
> - **Proximity Placement Groups:** For low-latency requirements.
> - **Regional Quotas:** vCPU quotas per family/region.
> - **Spot Eviction Policy:** Spot VMs have different eviction policies.
> - **Confidential Computing:** Support for confidential workloads.
> - **Ephemeral OS Disk:** Fast boot support.
> - **Availability Zones:** Not all SKUs are available in all zones.
> - **GPU/FPGA:** Specific GPU/FPGA types for workloads.
>
> These are modeled as fields and filter functions in the code, and can be extended as needed.
