#!/usr/bin/env python3
"""
Plots the scores of an algorithm in relation to the original demo time series (data/dataset.csv).

Use directly from the project root directory:
- `python scripts/plot-scores.py [results/scores.csv]`
- `./scripts/plot-scores.py [results/scores.csv]`
"""

import argparse
from typing import List

from matplotlib import axes
from matplotlib.patches import Patch
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

from pathlib import Path
from sklearn.preprocessing import MinMaxScaler

import scienceplots  # type: ignore # noqa: F401

plt.style.use(["science", "bright"])


def _create_arg_parser() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Plot time series, ground truth labels and anomaly scores"
    )
    parser.add_argument(
        "-d",
        "--data-file",
        type=Path,
        required=False,
        default="data/dataset.csv",
        help="File path to the dataset",
    )
    parser.add_argument(
        "-s",
        "--scores-file",
        type=Path,
        required=False,
        default="results/",
        help="Directory path to the scores",
    )
    parser.add_argument(
        "-i",
        "--ignore-label",
        action="store_true",
        required=False,
        help="Plot ground truth label",
    )
    parser.add_argument(
        "-n",
        "--name",
        type=str,
        required=True,
        help="Dataset name",
    )
    return parser.parse_args()


def plot(data, labels, scores, algorithms, name):
    axs: List[axes.Axes]
    fig, axs = plt.subplots(2, 1, sharex=True, figsize=(8.5, 4))

    # axs[0].set_title(f"Data from '{data_path}'")
    # axs[1].set_title(f"Scores from '{score_path}'")

    line_styles = ["-", "--", "-.", ":"]
    colors = ["b", "g", "r", "c", "m", "y", "k"]

    for i in range(data.shape[1]):
        axs[0].plot(
            data[:, 0],
            label="Time series",
            linestyle="-",
            color="blue",
        )

    # label all lines
    # axs[0].legend(frameon=True, loc="upper right")
    red_patch = Patch(facecolor="red", edgecolor="r", label="Anomaly", alpha=0.3)

    if labels is not None:
        # axs[1].plot(labels, label="ground truth", color="skyblue", linestyle="-.", alpha=0.5)
        x = np.arange(len(scores[0]))
        print(data[:, 0].shape)
        # axs[0].fill_between(np.arange(len(labels)),0,labels,where=(labels>0),color='skyblue',alpha=0.3)
        axs[0].fill_between(
            x, 0, np.max(data) * 1.1, where=labels[1:] == 1, color="red", alpha=0.3
        )
        # axs[0].fill_between(x,0,scores,where=labels[1:]==1,color='red',alpha=0.3)

    for i, score in enumerate(scores):
        axs[1].plot(score, label=algorithms[i])

    # axs[1].legend(frameon=True, loc="upper right")

    axs[0].spines["top"].set_visible(False)
    # axs[0].spines["right"].set_visible(False)
    axs[0].spines["bottom"].set_visible(False)
    axs[1].spines["top"].set_visible(False)
    # axs[1].spines["right"].set_visible(False)
    axs[0].tick_params(axis="x", which="both", bottom=False, top=False, right=False)
    axs[1].tick_params(axis="x", which="both", top=False, right=False)

    axs[0].set_ylabel("Value")
    axs[1].set_xlabel("Time", loc="center")
    axs[1].set_ylabel("Score")

    # Make sure at least three y ticks are shown
    axs[0].locator_params(axis="y", nbins=3)
    axs[1].locator_params(axis="y", nbins=3)
    # Show a grid on the y axis
    axs[0].grid(axis="y")
    axs[1].grid(axis="y")

    # set font size

    # add a line a 0.5
    # axs[1].axhline(0.5, color="red", linestyle="--", alpha=0.5)

    plt.tight_layout()
    # axs[0].set_xlim(0.0, np.max(data[:, 0]))

    # plt.show()
    # axs[0].set_xlim(0, 3400.0)
    lines = axs[0].get_lines()
    if name == "thermal":
        axs[0].legend(frameon=True, loc="upper left", handles=lines + [red_patch])
        axs[1].legend(frameon=True, loc="upper left")
    else:
        axs[0].legend(frameon=True, loc="upper right", handles=lines + [red_patch])
        axs[1].legend(frameon=True, loc="upper right")

    plt.savefig(f"{name}_all_plot.png", bbox_inches="tight")

    if name == "thermal":
        axs[0].set_xlim(2500, 3000.0)  # Therm
    elif name == "cpu":
        axs[0].set_xlim(7850.0, 10050.0)  # CPU
    elif name == "smb":
        axs[0].set_xlim(800, 1000)  # SMB

    axs[0].legend(frameon=True, loc="upper right", handles=lines + [red_patch])
    axs[1].legend(frameon=True, loc="upper right")

    plt.savefig(f"{name}_all_plot_zoom.png", bbox_inches="tight")


def main(data_path: Path, score_path: Path, plot_label: bool, name: str):
    print(f"Plotting data from '{data_path}' and scores from '{score_path}'")
    df = pd.read_csv(data_path)
    data = df.iloc[:, 1:-1].values
    labels = df.iloc[:, -1].values
    # Multiple scores
    scores = []
    algorithms = []
    # Find subdirectories of the scores directory
    for subdir in score_path.iterdir():
        if subdir.is_dir():
            # Find the score file in the subdirectory
            algorithm = subdir.name
            # Find the anomaly.ts file several subdirectories deep eg DWT-MLEAD/d751713988987e9331980363e24189ce/cpu-user/system-14_cpu-user/1/anomaly_scores.ts
            anomaly_ts = subdir.glob("**/anomaly_scores.ts")
            if not anomaly_ts:
                print(f"No anomaly scores found in {subdir}")
                continue
            anomaly_ts = list(anomaly_ts)[0]

            algorithms.append(algorithm)
            # Read the anomaly scores
            score = pd.read_csv(anomaly_ts).values
            scores.append(
                MinMaxScaler().fit_transform(score.reshape(-1, 1)).reshape(-1)
            )

    column_names = df.columns.values[1:-1]

    plot(data, labels if plot_label else None, scores, algorithms, name)


if __name__ == "__main__":
    args = _create_arg_parser()
    main(args.data_file, args.scores_file, not args.ignore_label, args.name)
