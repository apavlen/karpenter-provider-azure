package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/karpenter-provider-azure/pkg/resolver"
)

func main() {
	var (
		traceSource = flag.String("trace", "google", "Trace source: google|azure|alibaba|custom")
		skuFile     = flag.String("sku", "azure_skus.json", "Path to Azure SKU JSON file")
		maxRows     = flag.Int("max", 1000, "Max workloads to simulate")
		outFile     = flag.String("out", "", "Optional: output CSV file for results")
		workloadsFile = flag.String("workloads", "", "Optional: path to custom workloads JSON file")
	)
	flag.Parse()

	var src resolver.TraceSource
	switch *traceSource {
	case "google":
		src = resolver.TraceGoogle
	case "azure":
		src = resolver.TraceAzure
	case "alibaba":
		src = resolver.TraceAlibaba
	case "custom":
		src = resolver.TraceSource("custom")
	default:
		fmt.Fprintf(os.Stderr, "Unknown trace source: %s\n", *traceSource)
		os.Exit(1)
	}

	// If custom workloads file is provided, use it
	if src == "custom" && *workloadsFile != "" {
		result, naive, err := resolver.RunCustomWorkloadSimulation(*workloadsFile, *skuFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Simulation failed: %v\n", err)
			os.Exit(2)
		}
		if *outFile != "" {
			f, err := os.Create(*outFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
				os.Exit(3)
			}
			defer f.Close()
			fmt.Fprintf(f, "Strategy,VMs Used,Total Cost,Avg CPU Util (%),Avg Mem Util (%)\n")
			fmt.Fprintf(f, "NewAlgorithm,%d,%.2f,%.1f,%.1f\n", result.VMsUsed, result.TotalCost, result.AvgCPU, result.AvgMem)
			fmt.Fprintf(f, "Naive,%d,%.2f,%.1f,%.1f\n", naive.VMsUsed, naive.TotalCost, naive.AvgCPU, naive.AvgMem)
			fmt.Printf("Results written to %s\n", *outFile)
		}
		return
	}

	// Run simulation and capture results
	result, naive, err := resolver.RunTraceSimulationWithResults(src, *skuFile, *maxRows)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Simulation failed: %v\n", err)
		os.Exit(2)
	}

	// Optionally write results to CSV
	if *outFile != "" {
		f, err := os.Create(*outFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
			os.Exit(3)
		}
		defer f.Close()
		fmt.Fprintf(f, "Strategy,VMs Used,Total Cost,Avg CPU Util (%),Avg Mem Util (%)\n")
		fmt.Fprintf(f, "NewAlgorithm,%d,%.2f,%.1f,%.1f\n", result.VMsUsed, result.TotalCost, result.AvgCPU, result.AvgMem)
		fmt.Fprintf(f, "Naive,%d,%.2f,%.1f,%.1f\n", naive.VMsUsed, naive.TotalCost, naive.AvgCPU, naive.AvgMem)
		fmt.Printf("Results written to %s\n", *outFile)
	}
}
