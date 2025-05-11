import csv
import json
import os

# Paths to the raw Azure trace CSVs (update if needed)
VM_CHARACTERISTICS_CSV = "azure_traces/VM-characteristics.csv"
VM_USAGE_CSV = "azure_traces/VM-usage.csv"

# Output files for benchmarking
INSTANCE_SPECS_JSON = "azure_traces/instance_specs.json"
WORKLOAD_PROFILES_JSON = "azure_traces/workload_profiles.json"

def process_vm_characteristics():
    specs = []
    with open(VM_CHARACTERISTICS_CSV, newline='') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            spec = {
                "Name": row.get("vmSize", ""),
                "VCpus": int(row.get("vCPUs", "0")),
                "MemoryGiB": float(row.get("memoryGiB", "0")),
                "StorageGiB": float(row.get("maxDataDiskSizeGB", "0")),
                "PricePerHour": float(row.get("retailPrice", "0")),
                "Family": row.get("family", ""),
                "GPUCount": int(row.get("gpus", "0")),
                "GPUType": row.get("gpuType", ""),
                "AvailabilityZones": row.get("zones", "").split(";") if row.get("zones") else [],
                "EphemeralOSDisk": row.get("ephemeralOSDisk", "false").lower() == "true",
                "NestedVirtualization": row.get("nestedVirtualization", "false").lower() == "true",
                "SpotSupported": row.get("spot", "false").lower() == "true",
                "ConfidentialComputing": row.get("confidentialComputing", "false").lower() == "true",
                "TrustedLaunch": row.get("trustedLaunch", "false").lower() == "true",
                "AcceleratedNetworking": row.get("acceleratedNetworking", "false").lower() == "true",
                "MaxPods": int(row.get("maxPods", "0")),
                "UltraSSDEnabled": row.get("ultraSSDEnabled", "false").lower() == "true",
                "ProximityPlacement": row.get("proximityPlacement", "false").lower() == "true",
                "Capabilities": {},  # Can be filled in if needed
            }
            specs.append(spec)
    with open(INSTANCE_SPECS_JSON, "w") as f:
        json.dump(specs, f, indent=2)
    print(f"Wrote {len(specs)} instance specs to {INSTANCE_SPECS_JSON}")

def process_vm_usage():
    profiles = []
    with open(VM_USAGE_CSV, newline='') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            profile = {
                "CPURequirements": int(row.get("vCPUs", "0")),
                "MemoryRequirements": float(row.get("memoryGiB", "0")),
                "IORequirements": 0,  # Not present in trace, can be extended
                "GPURequirements": int(row.get("gpus", "0")),
                "GPUType": row.get("gpuType", ""),
                "Zone": row.get("zone", ""),
                "RequireEphemeralOS": False,  # Not present in trace
                "RequireNestedVirt": False,   # Not present in trace
                "RequireSpot": False,         # Not present in trace
                "RequireConfidential": False, # Not present in trace
                "Capabilities": {},           # Not present in trace
            }
            profiles.append(profile)
    with open(WORKLOAD_PROFILES_JSON, "w") as f:
        json.dump(profiles, f, indent=2)
    print(f"Wrote {len(profiles)} workload profiles to {WORKLOAD_PROFILES_JSON}")

if __name__ == "__main__":
    os.makedirs("azure_traces", exist_ok=True)
    process_vm_characteristics()
    process_vm_usage()
