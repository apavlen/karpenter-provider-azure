#!/usr/bin/env python3
"""
Plot simulation results from CSV output of the instance selection simulation CLI.

Usage:
  python3 scripts/plot_simulation_results.py results.csv

This will generate a bar chart comparing VMs used, total cost, and utilization for each strategy.
"""

import sys
import pandas as pd
import matplotlib.pyplot as plt

def main(csv_path):
    df = pd.read_csv(csv_path)
    strategies = df['Strategy']
    vms = df['VMs Used']
    cost = df['Total Cost']
    cpu = df['Avg CPU Util (%)']
    mem = df['Avg Mem Util (%)']

    fig, axs = plt.subplots(2, 2, figsize=(10, 8))
    fig.suptitle("Instance Selection Simulation Results")

    axs[0, 0].bar(strategies, vms, color='skyblue')
    axs[0, 0].set_title("VMs Used")
    axs[0, 0].set_ylabel("Count")

    axs[0, 1].bar(strategies, cost, color='orange')
    axs[0, 1].set_title("Total Cost ($/hr)")
    axs[0, 1].set_ylabel("Cost")

    axs[1, 0].bar(strategies, cpu, color='green')
    axs[1, 0].set_title("Avg CPU Utilization (%)")
    axs[1, 0].set_ylabel("Percent")

    axs[1, 1].bar(strategies, mem, color='purple')
    axs[1, 1].set_title("Avg Memory Utilization (%)")
    axs[1, 1].set_ylabel("Percent")

    for ax in axs.flat:
        ax.set_xticks(range(len(strategies)))
        ax.set_xticklabels(strategies)

    plt.tight_layout(rect=[0, 0.03, 1, 0.95])
    plt.show()

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 scripts/plot_simulation_results.py results.csv")
        sys.exit(1)
    main(sys.argv[1])
