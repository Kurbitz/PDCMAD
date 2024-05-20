#!/usr/bin/env python3
"""
Plots the scores of an algorithm in relation to the original demo time series (data/dataset.csv).

Use directly from the project root directory:
- `python scripts/plot-scores.py [results/scores.csv]`
- `./scripts/plot-scores.py [results/scores.csv]`
"""

import argparse
from typing import List

from matplotlib.patches import Patch
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from matplotlib import figure, axes
import scienceplots  # type: ignore # noqa: F401

plt.style.use(["science", "bright"])


from pathlib import Path
from sklearn.preprocessing import MinMaxScaler


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
        default="results/scores.csv",
        help="File path to the scores",
    )
    parser.add_argument(
        "-i",
        "--ignore-label",
        action="store_true",
        required=False,
        help="Plot ground truth label",
    )
    parser.add_argument(
        "-a",
        "--algorithm",
        type=str,
        required=False,
        default="",
        help="Algorithm name",
    )
    parser.add_argument(
        "-n",
        "--name",
        type=str,
        required=True,
        help="Dataset name",
    )

    return parser.parse_args()


def plot(data, labels, scores, algorithm, name):
    axs: List[axes.Axes]
    fig, axs = plt.subplots(2, 1, sharex=True, figsize=(8.5, 3))

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
    red_patch = Patch(facecolor="red", edgecolor="r", label="Anomaly", alpha=0.3)
    # get the line
    line = axs[0].get_lines()[0]

    if name == "thermal":
        axs[0].legend(handles=[line, red_patch], frameon=True, loc="upper left")
    else:
        axs[0].legend(handles=[line, red_patch], frameon=True, loc="upper right")

    if labels is not None:
        # axs[1].plot(labels, label="ground truth", color="skyblue", linestyle="-.", alpha=0.5)
        x = np.arange(len(scores))
        print(data[:, 0].shape)
        # axs[0].fill_between(np.arange(len(labels)),0,labels,where=(labels>0),color='skyblue',alpha=0.3)
        axs[0].fill_between(
            x, 0, np.max(data) * 1.1, where=labels[1:] == 1, color="red", alpha=0.3
        )
        # axs[0].fill_between(x,0,scores,where=labels[1:]==1,color='red',alpha=0.3)

    axs[1].plot(scores, label=algorithm, color="orange", linestyle="-")

    line = axs[0].get_lines()[0]
    if name == "thermal":
        axs[0].legend(frameon=True, loc="upper left", handles=[line, red_patch])
        axs[1].legend(frameon=True, loc="upper left")
    else:
        axs[0].legend(frameon=True, loc="upper right", handles=[line, red_patch])
        axs[1].legend(frameon=True, loc="upper right")

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

    plt.show()
    plt.savefig(f"{name}_{algorithm}_plot.png", bbox_inches="tight", format="png")

    if name == "smb":
        axs[0].set_xlim(800, 1000)  # SMB
    elif name == "cpu":
        axs[0].set_xlim(7850.0, 10050.0)  # CPU
    elif name == "thermal":
        axs[0].set_xlim(2500, 2920.0)

    axs[0].legend(frameon=True, loc="upper right", handles=[line, red_patch])
    axs[1].legend(frameon=True, loc="upper right")

    plt.savefig(f"{name}_{algorithm}_plot_zoom.png", bbox_inches="tight", format="png")


def main(
    data_path: Path, score_path: Path, plot_label: bool, algorithm: str, name: str
):
    print(f"Plotting data from '{data_path}' and scores from '{score_path}'")
    df = pd.read_csv(data_path)
    data = df.iloc[:, 1:-1].values
    labels = df.iloc[:, -1].values
    scores = pd.read_csv(score_path).values
    scores = MinMaxScaler().fit_transform(scores.reshape(-1, 1)).reshape(-1)

    plot(
        data,
        labels if plot_label else None,
        scores,
        algorithm,
        name,
    )


if __name__ == "__main__":
    args = _create_arg_parser()
    main(
        args.data_file,
        args.scores_file,
        not args.ignore_label,
        args.algorithm,
        args.name,
    )
