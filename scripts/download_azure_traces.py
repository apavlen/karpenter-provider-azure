#!/usr/bin/env python3
"""
Script to download a sample of Azure VM deployment and usage traces from the AzurePublicDataset.

- Downloads a representative sample of VM deployment data (from vm_deployments_aggregate_2020)
- Downloads matching VM CPU/memory usage data
- Handles large files efficiently (streaming/chunking)
- Saves in an easily accessible format (CSV or Parquet)

Usage:
    python scripts/download_azure_traces.py --outdir data/azure_traces

Requirements:
    pip install requests tqdm pandas

References:
    https://github.com/Azure/AzurePublicDataset/blob/master/AzurePublicDatasetLinksV2.txt
    https://github.com/Azure/AzurePublicDataset/tree/master/vm
"""

import os
import argparse
import requests
from tqdm import tqdm

# URLs for the 2020 VM deployment and usage data (update if needed)
# NOTE: As of 2024, these direct links may no longer work or may require special access.
# Microsoft has restricted public access to the full Azure VM traces due to privacy and compliance.
# To request access or get the latest instructions, visit:
#   https://github.com/Azure/AzurePublicDataset
#   https://github.com/Azure/AzurePublicDataset/blob/master/AzurePublicDatasetLinksV2.txt
#   https://www.microsoft.com/en-us/research/project/azure-vm-placement-trace/
#
# Steps to get the traces:
# 1. Go to the AzurePublicDataset GitHub repo and read the README and issues for current access instructions.
# 2. If required, fill out the request form or email Microsoft Research as described in the repo.
# 3. Once you have access, download the files manually or update the URLs below with your authorized links.
# 4. Place the files in the expected directory (e.g., data/azure_traces/) for use with the preprocessing script.

VM_DEPLOYMENTS_URL = "https://azureopendatastorage.blob.core.windows.net/azurepublicdataset/vm_deployments_aggregate_2020.csv"
VM_USAGE_URL = "https://azureopendatastorage.blob.core.windows.net/azurepublicdataset/vm_cpu_mem_2020.csv"

def download_file(url, out_path, chunk_size=1024*1024):
    """Download a file with progress bar and chunking. Handles 404 and XML error responses."""
    response = requests.get(url, stream=True)
    if response.status_code != 200 or response.headers.get('content-type', '').startswith('application/xml'):
        print(f"ERROR: Failed to download {url} (status {response.status_code})")
        print("Response headers:", response.headers)
        print("Response text:", response.text[:500])
        raise RuntimeError(f"Failed to download {url}")
    total = int(response.headers.get('content-length', 0))
    with open(out_path, 'wb') as file, tqdm(
        desc=f"Downloading {os.path.basename(out_path)}",
        total=total, unit='B', unit_scale=True
    ) as bar:
        for chunk in response.iter_content(chunk_size=chunk_size):
            if chunk:
                file.write(chunk)
                bar.update(len(chunk))
    # After download, print the first few lines to help debug
    print(f"First 3 lines of {out_path}:")
    with open(out_path, "r", encoding="utf-8", errors="replace") as f:
        for i in range(3):
            line = f.readline()
            if not line:
                break
            print(line.strip())
    # If the file starts with "<?xml", warn the user
    with open(out_path, "r", encoding="utf-8", errors="replace") as f:
        first_line = f.readline()
        if first_line.startswith("<?xml"):
            print(f"WARNING: {out_path} appears to be an XML error file, not a CSV.")
            print("Please check the AzurePublicDataset repo for updated links or access instructions.")
            print("See: https://github.com/Azure/AzurePublicDataset")

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--outdir', type=str, default='data/azure_traces', help='Output directory')
    args = parser.parse_args()

    os.makedirs(args.outdir, exist_ok=True)

    # Download VM deployments
    deployments_path = os.path.join(args.outdir, "vm_deployments_aggregate_2020.csv")
    if not os.path.exists(deployments_path):
        print(
            "\nNOTE: If this download fails or the file is XML, you likely need to request access to the Azure VM traces.\n"
            "See the comments at the top of this script for instructions.\n"
        )
        download_file(VM_DEPLOYMENTS_URL, deployments_path)
    else:
        print(f"File already exists: {deployments_path}")

    # Download VM usage
    usage_path = os.path.join(args.outdir, "vm_cpu_mem_2020.csv")
    if not os.path.exists(usage_path):
        print(
            "\nNOTE: If this download fails or the file is XML, you likely need to request access to the Azure VM traces.\n"
            "See the comments at the top of this script for instructions.\n"
        )
        download_file(VM_USAGE_URL, usage_path)
    else:
        print(f"File already exists: {usage_path}")

    print("Download complete.")

if __name__ == "__main__":
    main()
