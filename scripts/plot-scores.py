#!/usr/bin/env python3
"""
Plots the scores of an algorithm in relation to the original demo time series (data/dataset.csv).

Use directly from the project root directory:
- `python scripts/plot-scores.py [results/scores.csv]`
- `./scripts/plot-scores.py [results/scores.csv]`
"""

import argparse

import pandas as pd
import matplotlib.pyplot as plt

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
    return parser.parse_args()


def plot(data, labels, scores, column_names, data_path, score_path):
    fig, axs = plt.subplots(2, 1, sharex=True, figsize=(20, 10))

    axs[0].set_title(f"Data from '{data_path}'")
    axs[1].set_title(f"Scores from '{score_path}'")

    line_styles = ["-", "--", "-.", ":"]
    colors = ["b", "g", "r", "c", "m", "y", "k"]

    for i in range(data.shape[1]):
        axs[0].plot(
            data[:, i],
            label=column_names[i],
            linestyle=line_styles[i % len(line_styles)],
            color=colors[i % len(colors)],
        )

    # label all lines
    axs[0].legend()

    if labels is not None:
        axs[1].plot(labels, label="ground truth", color="blue", linestyle="-.")
    axs[1].plot(scores, label="score", color="orange", linestyle="--")
    axs[1].legend("ground truth", "score")

    # add a line a 0.5
    axs[1].axhline(0.5, color="red", linestyle="--")

    plt.legend()

    plt.show()


def main(data_path: Path, score_path: Path, plot_label: bool):
    print(f"Plotting data from '{data_path}' and scores from '{score_path}'")
    df = pd.read_csv(data_path)
    data = df.iloc[:, 1:-1].values
    labels = df.iloc[:, -1].values
    scores = pd.read_csv(score_path).values
    scores = MinMaxScaler().fit_transform(scores.reshape(-1, 1)).reshape(-1)

    column_names = df.columns.values[1:-1]

    plot(
        data,
        labels if plot_label else None,
        scores,
        column_names,
        data_path,
        score_path,
    )


if __name__ == "__main__":
    args = _create_arg_parser()
    main(args.data_file, args.scores_file, not args.ignore_label)
