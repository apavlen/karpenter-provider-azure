# Azure Traces for Benchmarking

## Download Instructions

1. Download the following files from the [AzurePublicDataset releases page](https://github.com/Azure/AzurePublicDataset/releases):

   - `VM-characteristics.csv`
   - `VM-usage.csv`

2. Place both files in the `azure_traces/` directory at the root of your project.

## Processing Instructions

Run the following command to process the raw CSVs into JSON files for benchmarking:

```bash
python3 scripts/process_azure_traces.py
```

This will generate:

- `azure_traces/instance_specs.json` (list of AzureInstanceSpec)
- `azure_traces/workload_profiles.json` (list of WorkloadProfile)

You can now use these JSON files as input for your Go benchmarks or simulations.
