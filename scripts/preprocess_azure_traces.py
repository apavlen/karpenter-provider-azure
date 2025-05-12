#!/usr/bin/env python3
"""
Preprocess Azure VM deployment and usage traces into Kubernetes pod-like workload profiles.

- Automatically downloads Azure trace files if not present
- Converts VM sizes to CPU/memory requests (with comprehensive mapping)
- Handles large files efficiently (streaming, chunking, and optional row limit for debugging)
- Flexible column name detection
- Robust error handling and logging

Usage:
    python scripts/preprocess_azure_traces.py --indir data/azure_traces --out workloads_preprocessed.json [--limit 100000]

Requirements:
    pip install pandas tqdm requests

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
import requests
import sys
import glob

# Correct file names for Azure Public Dataset v2
DEPLOYMENTS_FILE = "trace_data/deployments/deployments.csv.gz"
USAGE_GLOB = "trace_data/vm_cpu_readings/vm_cpu_readings-file-*.csv.gz"

# Comprehensive mapping from Azure VM size to (vCPUs, MemoryGiB)
AZURE_VM_SIZE_MAP = {
    # Dv3
    "Standard_D2_v3": (2, 8), "Standard_D4_v3": (4, 16), "Standard_D8_v3": (8, 32),
    "Standard_D16_v3": (16, 64), "Standard_D32_v3": (32, 128), "Standard_D64_v3": (64, 256),
    # Dsv3
    "Standard_D2s_v3": (2, 8), "Standard_D4s_v3": (4, 16), "Standard_D8s_v3": (8, 32),
    "Standard_D16s_v3": (16, 64), "Standard_D32s_v3": (32, 128), "Standard_D64s_v3": (64, 256),
    # Ev3
    "Standard_E2_v3": (2, 16), "Standard_E4_v3": (4, 32), "Standard_E8_v3": (8, 64),
    "Standard_E16_v3": (16, 128), "Standard_E32_v3": (32, 256), "Standard_E64_v3": (64, 432),
    # Esv3
    "Standard_E2s_v3": (2, 16), "Standard_E4s_v3": (4, 32), "Standard_E8s_v3": (8, 64),
    "Standard_E16s_v3": (16, 128), "Standard_E32s_v3": (32, 256), "Standard_E64s_v3": (64, 432),
    # Fsv2
    "Standard_F2s_v2": (2, 4), "Standard_F4s_v2": (4, 8), "Standard_F8s_v2": (8, 16),
    "Standard_F16s_v2": (16, 32), "Standard_F32s_v2": (32, 64), "Standard_F64s_v2": (64, 128), "Standard_F72s_v2": (72, 144),
    # B-series
    "Standard_B1s": (1, 1), "Standard_B2s": (2, 4), "Standard_B1ms": (1, 2), "Standard_B2ms": (2, 8),
    "Standard_B4ms": (4, 16), "Standard_B8ms": (8, 32),
    # A-series
    "Standard_A1_v2": (1, 2), "Standard_A2_v2": (2, 4), "Standard_A4_v2": (4, 8), "Standard_A8_v2": (8, 16),
    # Add more as needed for your dataset
}

def log(msg):
    print(f"[preprocess] {msg}", file=sys.stderr)

def download_file(url, out_path, chunk_size=1024*1024):
    log(f"Downloading {url} to {out_path}")
    try:
        with requests.get(url, stream=True, timeout=60) as r:
            r.raise_for_status()
            total = int(r.headers.get('content-length', 0))
            with open(out_path, 'wb') as f, tqdm(
                desc=f"Downloading {os.path.basename(out_path)}",
                total=total, unit='B', unit_scale=True
            ) as bar:
                for chunk in r.iter_content(chunk_size=chunk_size):
                    if chunk:
                        f.write(chunk)
                        bar.update(len(chunk))
        # Check for XML error
        with open(out_path, "r", encoding="utf-8", errors="replace") as f:
            first_line = f.readline()
            if first_line.startswith("<?xml"):
                log(f"ERROR: {out_path} appears to be an XML error file, not a CSV.")
                log("Please check the AzurePublicDataset repo for updated links or access instructions.")
                raise RuntimeError(f"Downloaded file {out_path} is not a valid CSV.")
    except Exception as e:
        log(f"Failed to download {url}: {e}")
        raise

def ensure_file(path, url):
    if not os.path.exists(path):
        log(f"{path} not found, downloading from {url}")
        download_file(url, path)
    else:
        log(f"File already exists: {path}")

def detect_column(df, expected_names):
    """
    Find the best matching column in df for each expected name (case-insensitive, underscore/space/strip).
    Returns a dict: expected_name -> actual_column_name
    """
    actual = [c.lower().replace("_", "").replace(" ", "") for c in df.columns]
    mapping = {}
    for exp in expected_names:
        exp_norm = exp.lower().replace("_", "").replace(" ", "")
        found = None
        for i, act in enumerate(actual):
            if act == exp_norm:
                found = df.columns[i]
                break
        if not found:
            # Try substring match
            for i, act in enumerate(actual):
                if exp_norm in act:
                    found = df.columns[i]
                    break
        if found:
            mapping[exp] = found
        else:
            mapping[exp] = None
    return mapping

def load_vm_size_map():
    # In production, load a full mapping from Azure docs or API
    return AZURE_VM_SIZE_MAP

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--indir', type=str, default='trace_data', help='Input directory (root of Azure trace)')
    parser.add_argument('--out', type=str, default='workloads_preprocessed.json', help='Output file')
    parser.add_argument('--limit', type=int, default=None, help='Limit number of rows (for debugging)')
    args = parser.parse_args()

    # File paths
    deployments_path = os.path.join(args.indir, "deployments", "deployments.csv.gz")
    usage_glob = os.path.join(args.indir, "vm_cpu_readings", "vm_cpu_readings-file-*.csv.gz")

    # Check files exist
    if not os.path.exists(deployments_path):
        log(f"ERROR: {deployments_path} not found. Please extract the Azure trace dataset as per documentation.")
        sys.exit(1)
    usage_files = sorted(glob.glob(usage_glob))
    if not usage_files:
        log(f"ERROR: No usage files found matching {usage_glob}")
        sys.exit(1)

    # Load deployments
    log("Loading VM deployments...")
    deployments = pd.read_csv(deployments_path, compression="gzip", nrows=args.limit)
    log(f"Deployment columns: {deployments.columns.tolist()}")
    expected_deploy_cols = ["vm_id", "vm_size", "start_time", "end_time", "resource_group"]
    deploy_col_map = detect_column(deployments, expected_deploy_cols)
    missing = [k for k, v in deploy_col_map.items() if v is None]
    if missing:
        log(f"Missing columns in deployments file: {missing}")
        sys.exit(1)
    deployments = deployments[[deploy_col_map[c] for c in expected_deploy_cols]]
    deployments.columns = expected_deploy_cols

    # Load and concatenate all usage files
    log("Loading and concatenating VM usage files...")
    usage_chunks = []
    total_rows = 0
    for fpath in usage_files:
        log(f"Reading {fpath}")
        chunk = pd.read_csv(fpath, compression="gzip", nrows=args.limit)
        usage_chunks.append(chunk)
        total_rows += len(chunk)
        if args.limit and total_rows >= args.limit:
            break
    usage = pd.concat(usage_chunks, ignore_index=True)
    if args.limit:
        usage = usage.iloc[:args.limit]
    log(f"Usage columns: {usage.columns.tolist()}")
    expected_usage_cols = ["vm_id", "timestamp", "cpu_usage", "mem_usage"]
    usage_col_map = detect_column(usage, expected_usage_cols)
    missing_usage = [k for k, v in usage_col_map.items() if v is None]
    if missing_usage:
        log(f"Missing columns in usage file: {missing_usage}")
        sys.exit(1)
    usage = usage[[usage_col_map[c] for c in expected_usage_cols]]
    usage.columns = expected_usage_cols

    vm_size_map = load_vm_size_map()

    # Merge deployments and usage on vm_id
    log("Merging deployments and usage data...")
    merged = pd.merge(usage, deployments, on="vm_id", how="inner")

    # Filter: only keep usage records within the VM's lifetime
    log("Filtering usage records to VM lifetime...")
    merged = merged[
        (merged["timestamp"] >= merged["start_time"]) &
        (merged["timestamp"] <= merged["end_time"])
    ]

    # Deduplicate: keep only one usage record per (vm_id, timestamp)
    merged = merged.drop_duplicates(subset=["vm_id", "timestamp"])

    # Convert to pod-like workload profiles
    log("Converting merged data to pod-like workload profiles...")
    workloads = []
    for _, row in tqdm(merged.iterrows(), total=len(merged)):
        vm_size = row["vm_size"]
        vcpus, mem_gib = vm_size_map.get(vm_size, (None, None))
        if vcpus is None:
            # Try to parse vCPU/mem from VM size string (e.g., Standard_D4_v3 -> 4 vCPUs, 16 GiB)
            import re
            m = re.match(r"Standard_[A-Z]+(\d+)[a-z_]*_v\d+", str(vm_size))
            if m:
                try:
                    vcpus = int(m.group(1))
                    # Heuristic: 1 vCPU = 4 GiB (for D-series), fallback if not mapped
                    mem_gib = vcpus * 4
                except Exception:
                    vcpus, mem_gib = None, None
        if vcpus is None:
            log(f"Skipping unknown VM size: {vm_size}")
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

    log(f"Writing {len(workloads)} workloads to {args.out}")
    with open(args.out, "w") as f:
        json.dump(workloads, f, indent=2)

if __name__ == "__main__":
    main()
