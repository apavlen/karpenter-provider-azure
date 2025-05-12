#!/usr/bin/env python3
"""
Visualize and analyze benchmark results for Azure instance selection strategies.

- Loads benchmark results (JSON or CSV)
- Plots comparative metrics: cost efficiency, resource utilization, instance diversity, selection time
- Identifies patterns where specific strategies perform best
- Calculates efficiency gains compared to baseline

Usage:
    python scripts/visualize_benchmark_results.py --results benchmark_results.json

Requirements:
    pip install pandas matplotlib seaborn

Input:
    - benchmark_results.json (or .csv)

Output:
    - benchmark_results.html (optional, for interactive plots)
"""

import argparse
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--results', type=str, required=True, help='Benchmark results file (JSON or CSV)')
    args = parser.parse_args()

    # Load results
    if args.results.endswith('.json'):
        df = pd.read_json(args.results)
    else:
        df = pd.read_csv(args.results)

    # Example: Plot cost efficiency by strategy
    plt.figure(figsize=(8, 5))
    sns.barplot(data=df, x='strategy', y='cost_efficiency')
    plt.title('Cost Efficiency by Strategy')
    plt.ylabel('Estimated Monthly Cost ($)')
    plt.xlabel('Selection Strategy')
    plt.tight_layout()
    plt.show()

    # Example: Plot resource utilization
    plt.figure(figsize=(8, 5))
    sns.barplot(data=df, x='strategy', y='resource_utilization')
    plt.title('Resource Utilization by Strategy')
    plt.ylabel('CPU/Memory Utilization Ratio')
    plt.xlabel('Selection Strategy')
    plt.tight_layout()
    plt.show()

    # Example: Plot instance diversity
    plt.figure(figsize=(8, 5))
    sns.barplot(data=df, x='strategy', y='instance_diversity')
    plt.title('Instance Diversity by Strategy')
    plt.ylabel('Unique VM Types Selected')
    plt.xlabel('Selection Strategy')
    plt.tight_layout()
    plt.show()

    # Example: Plot selection time
    plt.figure(figsize=(8, 5))
    sns.barplot(data=df, x='strategy', y='selection_time_ms')
    plt.title('Selection Time by Strategy')
    plt.ylabel('Selection Time (ms)')
    plt.xlabel('Selection Strategy')
    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    main()
