package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Azure/karpenter-provider-azure/pkg/resolver"
)

func main() {
	var (
		traceSource = flag.String("trace", "google", "Trace source: google|azure|alibaba")
		skuFile     = flag.String("sku", "azure_skus.json", "Path to Azure SKU JSON file")
		maxRows     = flag.Int("max", 1000, "Max workloads to simulate")
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
	default:
		fmt.Fprintf(os.Stderr, "Unknown trace source: %s\n", *traceSource)
		os.Exit(1)
	}

	if err := resolver.RunTraceSimulation(src, *skuFile, *maxRows); err != nil {
		fmt.Fprintf(os.Stderr, "Simulation failed: %v\n", err)
		os.Exit(2)
	}
}
