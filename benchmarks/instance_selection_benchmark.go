package main

// Note: If you experience "ModuleNotFoundError: No module named 'apt_pkg'",
// this is likely due to an environment issue with Python's command-not-found handling.
// Consider installing python3-apt (sudo apt-get install python3-apt) or adjusting your environment.

import (
    "encoding/csv"
    "fmt"
    "math/rand"
    "os"
    "strconv"
    "time"

    "github.com/Azure/karpenter-provider-azure/pkg/resolver"
)

func main() {
    rand.Seed(time.Now().UnixNano())

    // Load workload profiles from CSV if available, otherwise generate synthetic workloads.
    var workloads []resolver.WorkloadProfile
    file, err := os.Open("./traces/sample_trace.csv")
    if err == nil {
        defer file.Close()
        reader := csv.NewReader(file)
        records, err := reader.ReadAll()
        if err == nil {
            for i, record := range records {
                if i == 0 {
                    // Skip header if present.
                    continue
                }
                cpuVal, _ := strconv.Atoi(record[1])
                memVal, _ := strconv.Atoi(record[2])
                workloads = append(workloads, resolver.WorkloadProfile{
                    CPURequirements:    cpuVal,
                    MemoryRequirements: memVal,
                })
            }
            fmt.Printf("Loaded %d workloads from CSV file.\n", len(workloads))
        } else {
            fmt.Println("Error reading CSV, generating synthetic workloads.")
        }
    }
    if len(workloads) == 0 {
        workloads = make([]resolver.WorkloadProfile, 1000)
        for i := 0; i < 1000; i++ {
            workloads[i] = resolver.WorkloadProfile{
                CPURequirements:    rand.Intn(8) + 1,
                MemoryRequirements: rand.Intn(32) + 1,
            }
        }
        fmt.Printf("Generated %d synthetic workloads.\n", len(workloads))
    }

    // Define candidate VMs.
    candidates := []resolver.VMType{
        {
            PricePerVCpu:   0.1,
            PricePerGiB:    0.2,
            VCpuCapacity:   8,
            MemoryCapacity: 32,
        },
        {
            PricePerVCpu:   0.2,
            PricePerGiB:    0.25,
            VCpuCapacity:   4,
            MemoryCapacity: 16,
        },
        // add additional candidate VMs as needed.
    }

    start := time.Now()
    for _, workload := range workloads {
        _ = resolver.SelectBestVM(candidates, workload)
    }
    elapsed := time.Since(start)
    fmt.Printf("Processed %d selections in %v\n", len(workloads), elapsed)
}
