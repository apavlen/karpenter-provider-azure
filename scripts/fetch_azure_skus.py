#!/usr/bin/env python3
"""
Fetch Azure VM SKU data and output as JSON for simulation.

Usage:
  python3 scripts/fetch_azure_skus.py > azure_skus.json
"""

import json
import requests

# Azure VM sizes API (public, no auth required for this endpoint)
API = "https://prices.azure.com/api/retail/prices?$filter=serviceName eq 'Virtual Machines' and armRegionName eq 'eastus'"

def fetch_all_skus():
    skus = []
    url = API
    while url:
        print(f"Fetching {url} ...")
        resp = requests.get(url)
        data = resp.json()
        for item in data.get("Items", []):
            if item.get("productName") and "vCPU" in item.get("productName"):
                # crude filter, refine as needed
                skus.append(item)
        url = data.get("NextPageLink")
    return skus

def to_sim_format(skus):
    # Convert Azure API format to simulation format
    out = []
    for s in skus:
        try:
            out.append({
                "Name": s["skuName"],
                "VCpus": int(s.get("cores", 0)),
                "MemoryGiB": float(s.get("memory", 0)),
                "StorageGiB": float(s.get("diskSizeGB", 0)),
                "PricePerHour": float(s.get("retailPrice", 0)),
                "Family": s.get("productFamily", ""),
                "Capabilities": {},
                "GPUCount": 0,
                "GPUType": "",
                "AvailabilityZones": [],
                "EphemeralOSDisk": False,
                "NestedVirtualization": False,
                "SpotSupported": "Spot" in s.get("skuName", ""),
                "ConfidentialComputing": False,
                "TrustedLaunch": False,
                "AcceleratedNetworking": False,
                "MaxPods": 0,
                "UltraSSDEnabled": False,
                "ProximityPlacement": False,
            })
        except Exception:
            continue
    return out

if __name__ == "__main__":
    skus = fetch_all_skus()
    sim_skus = to_sim_format(skus)
    print(json.dumps(sim_skus, indent=2))
