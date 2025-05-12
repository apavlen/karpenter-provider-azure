#!/usr/bin/env python3
"""
Preprocess Azure Public Dataset vmtable.csv.gz into Kubernetes pod-like workload profiles.

- Reads vmtable.csv.gz (downloaded by user) from --indir
- Extracts fields: vm_id, start_time, end_time, cpu_avg, mem_avg, vcpus, memory_gib, workload_type
- Converts start_time/end_time from int seconds to ISO8601 timestamps (base: 2020-01-01T00:00:00Z)
- Outputs pod-like workload dicts as JSON

Usage:
    python scripts/preprocess_azure_traces.py --indir azure_traces --out workloads_preprocessed.json [--limit 100000]

Requirements:
    pip install pandas tqdm

Input file:
    - vmtable.csv.gz (must be present in --indir)

Output:
    - workloads_preprocessed.json (list of pod-like workload dicts)
"""

import os
import argparse
import pandas as pd
import json
from tqdm import tqdm
import sys
from datetime import datetime, timedelta

VMTABLE_FILENAME = "vmtable.csv.gz"
VMTABLE_URL = "https://azurepublicdatasettraces.blob.core.windows.net/azurepublicdatasetv2/trace_data/vmtable/vmtable.csv.gz"

def log(msg):
    print(f"[preprocess] {msg}", file=sys.stderr)

def find_vmtable(indir):
    candidates = [
        os.path.join(indir, VMTABLE_FILENAME),
        os.path.join(indir, "vmtable", VMTABLE_FILENAME),
        os.path.join(indir, "trace_data", "vmtable", VMTABLE_FILENAME),
        os.path.join(indir, "AzurePublicDataset-master", "trace_data", "vmtable", VMTABLE_FILENAME),
        os.path.join(indir, "AzurePublicDataset-master", "vmtable", VMTABLE_FILENAME),
    ]
    for candidate in candidates:
        if os.path.exists(candidate):
            return candidate
    return None

def convert_time(secs):
    # Azure trace base time: 2020-01-01T00:00:00Z
    try:
        base = datetime(2020, 1, 1, 0, 0, 0)
        return (base + timedelta(seconds=int(secs))).strftime("%Y-%m-%dT%H:%M:%SZ")
    except Exception:
        return None

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--indir', type=str, default='azure_traces', help='Input directory (where vmtable.csv.gz is located)')
    parser.add_argument('--out', type=str, default='workloads_preprocessed.json', help='Output file')
    parser.add_argument('--limit', type=int, default=None, help='Limit number of rows (for debugging)')
    args = parser.parse_args()

    vmtable_path = find_vmtable(args.indir)
    if not vmtable_path:
        log("ERROR: Could not find vmtable.csv.gz in any known location. Searched:")
        log(f"  {os.path.join(args.indir, VMTABLE_FILENAME)} and related subdirs")
        log(f"Please download vmtable.csv.gz from:\n  {VMTABLE_URL}\nand place it in the --indir directory.")
        sys.exit(1)

    log(f"Loading vmtable from {vmtable_path} ...")
    try:
        df = pd.read_csv(vmtable_path, compression="gzip", nrows=args.limit)
    except Exception as e:
        log(f"ERROR: Failed to read {vmtable_path}: {e}")
        sys.exit(1)

    expected_cols = ["vm_id", "start_time", "end_time", "cpu_avg", "mem_avg", "vcpus", "memory_gib", "workload_type"]
    missing = [col for col in expected_cols if col not in df.columns]
    if missing:
        log(f"ERROR: Missing columns in vmtable: {missing}")
        log(f"Columns found: {df.columns.tolist()}")
        sys.exit(1)

    log("Converting vmtable rows to pod-like workload profiles...")
    workloads = []
    for _, row in tqdm(df.iterrows(), total=len(df)):
        start_ts = convert_time(row["start_time"])
        end_ts = convert_time(row["end_time"])
        if start_ts is None or end_ts is None:
            log(f"Skipping row with invalid start/end time: vm_id={row['vm_id']}")
            continue
        workload = {
            "name": f"workload-{row['vm_id']}",
            "cpu_request": int(row["vcpus"]),
            "memory_request_gib": float(row["memory_gib"]),
            "cpu_usage": float(row["cpu_avg"]),
            "mem_usage": float(row["mem_avg"]),
            "start_time": start_ts,
            "end_time": end_ts,
            "labels": {
                "workload_type": str(row["workload_type"]),
            },
            "annotations": {
                "azure_vm_id": str(row["vm_id"]),
            }
        }
        workloads.append(workload)

    log(f"Writing {len(workloads)} workloads to {args.out}")
    try:
        with open(args.out, "w") as f:
            json.dump(workloads, f, indent=2)
    except Exception as e:
        log(f"ERROR: Failed to write output file: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
