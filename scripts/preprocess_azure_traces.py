#!/usr/bin/env python3
"""
Preprocess Azure VM deployment and usage traces into Kubernetes pod-like workload profiles.

- Converts VM sizes to CPU/memory requests
- Creates synthetic pod specifications
- Preserves temporal patterns from the original data
- Generates labels/annotations for testing constraints

Usage:
    python scripts/preprocess_azure_traces.py --indir data/azure_traces --out workloads_preprocessed.json

Requirements:
    pip install pandas tqdm

Input files:
    - vm_deployments_aggregate_2020.csv
    - vm_cpu_mem_2020.csv

Output:
    - workloads_preprocessed.json (list of pod-like workload dicts)
"""

import os
import argparse
import pandas as pd
import json
from tqdm import tqdm

# Example mapping from Azure VM size to CPU/memory (can be extended)
AZURE_VM_SIZE_MAP = {
    # VMSize: (vCPUs, MemoryGiB)
    "Standard_D2_v3": (2, 8),
    "Standard_D4_v3": (4, 16),
    "Standard_D8_v3": (8, 32),
    "Standard_D16_v3": (16, 64),
    "Standard_D32_v3": (32, 128),
    # Add more as needed
}

def load_vm_size_map():
    # In production, load a full mapping from Azure docs or API
    return AZURE_VM_SIZE_MAP

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--indir', type=str, default='data/azure_traces', help='Input directory')
    parser.add_argument('--out', type=str, default='workloads_preprocessed.json', help='Output file')
    args = parser.parse_args()

    deployments_path = os.path.join(args.indir, "vm_deployments_aggregate_2020.csv")
    usage_path = os.path.join(args.indir, "vm_cpu_mem_2020.csv")

    print("Loading VM deployments...")
    # Read without usecols to inspect columns if needed
    deployments = pd.read_csv(deployments_path, nrows=100000)
    print("Deployment columns:", deployments.columns.tolist())
    # Check for XML error (AzurePublicDataset sometimes returns XML error if file not found)
    if len(deployments.columns) == 1 and deployments.columns[0].startswith("<?xml"):
        print("ERROR: The deployments CSV file does not contain valid data. It may be an error XML.")
        print("File content preview:", deployments.columns[0][:200])
        print("Please check that the file was downloaded correctly and the URL is valid.")
        exit(1)
    # Try to find the correct column names (case-insensitive)
    expected_cols = ["vm_id", "vm_size", "start_time", "end_time", "resource_group"]
    actual_cols = [c.lower() for c in deployments.columns]
    col_map = {col: deployments.columns[actual_cols.index(col)] for col in expected_cols if col in actual_cols}
    missing = [col for col in expected_cols if col not in actual_cols]
    if missing:
        raise ValueError(f"Missing columns in deployments file: {missing}")
    deployments = deployments[[col_map[c] for c in expected_cols]]
    deployments.columns = expected_cols

    print("Loading VM usage...")
    usage = pd.read_csv(usage_path, nrows=100000)
    print("Usage columns:", usage.columns.tolist())
    expected_usage_cols = ["vm_id", "timestamp", "cpu_usage", "mem_usage"]
    actual_usage_cols = [c.lower() for c in usage.columns]
    usage_col_map = {col: usage.columns[actual_usage_cols.index(col)] for col in expected_usage_cols if col in actual_usage_cols}
    missing_usage = [col for col in expected_usage_cols if col not in actual_usage_cols]
    if missing_usage:
        raise ValueError(f"Missing columns in usage file: {missing_usage}")
    usage = usage[[usage_col_map[c] for c in expected_usage_cols]]
    usage.columns = expected_usage_cols

    vm_size_map = load_vm_size_map()

    # Merge deployments and usage on vm_id
    merged = pd.merge(usage, deployments, on="vm_id", how="inner")

    # Convert to pod-like workload profiles
    workloads = []
    for _, row in tqdm(merged.iterrows(), total=len(merged)):
        vm_size = row["vm_size"]
        vcpus, mem_gib = vm_size_map.get(vm_size, (None, None))
        if vcpus is None:
            continue  # skip unknown sizes

        # Synthesize a pod spec
        workload = {
            "name": f"workload-{row['vm_id']}-{row['timestamp']}",
            "cpu_request": vcpus,
            "memory_request_gib": mem_gib,
            "cpu_usage": row["cpu_usage"],
            "mem_usage": row["mem_usage"],
            "start_time": row["start_time"],
            "end_time": row["end_time"],
            "labels": {
                "resource_group": row["resource_group"],
                "vm_size": vm_size,
            },
            "annotations": {
                "azure_vm_id": row["vm_id"],
            }
        }
        workloads.append(workload)

    print(f"Writing {len(workloads)} workloads to {args.out}")
    with open(args.out, "w") as f:
        json.dump(workloads, f, indent=2)

if __name__ == "__main__":
    main()
