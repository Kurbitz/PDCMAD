import json
import time
import sys
import csv
import argparse
import enum
from typing import Any, List, Tuple
import math
import itertools


class LttbException(Exception):
    pass


def largest_triangle_three_buckets(data, threshold) -> List[Tuple[int, float]]:
    """
    Return a downsampled version of data.
    Parameters
    ----------
    data: list of lists/tuples
        data must be formated this way: [[x,y], [x,y], [x,y], ...]
                                    or: [(x,y), (x,y), (x,y), ...]
    threshold: int
        threshold must be >= 2 and <= to the len of data
    Returns
    -------
    data, but downsampled using threshold
    """

    # Check if data and threshold are valid
    if not isinstance(data, list):
        raise LttbException("data is not a list")
    if not isinstance(threshold, int) or threshold <= 2 or threshold >= len(data):
        raise LttbException("threshold not well defined")
    for i in data:
        if not isinstance(i, (list, tuple)) or len(i) != 2:
            raise LttbException("datapoints are not lists or tuples")

    # Bucket size. Leave room for start and end data points
    every = (len(data) - 2) / (threshold - 2)

    a = 0  # Initially a is the first point in the triangle
    next_a = 0
    max_area_point = (0, 0)

    sampled = [data[0]]  # Always add the first point

    for i in range(0, threshold - 2):
        # Calculate point average for next bucket (containing c)
        avg_x = 0
        avg_y = 0
        avg_range_start = int(math.floor((i + 1) * every) + 1)
        avg_range_end = int(math.floor((i + 2) * every) + 1)
        avg_rang_end = avg_range_end if avg_range_end < len(data) else len(data)

        avg_range_length = avg_rang_end - avg_range_start

        while avg_range_start < avg_rang_end:
            avg_x += data[avg_range_start][0]
            avg_y += data[avg_range_start][1]
            avg_range_start += 1

        avg_x /= avg_range_length
        avg_y /= avg_range_length

        # Get the range for this bucket
        range_offs = int(math.floor((i + 0) * every) + 1)
        range_to = int(math.floor((i + 1) * every) + 1)

        # Point a
        point_ax = data[a][0]
        point_ay = data[a][1]

        max_area = -1

        while range_offs < range_to:
            # Calculate triangle area over three buckets
            area = (
                math.fabs(
                    (point_ax - avg_x) * (data[range_offs][1] - point_ay)
                    - (point_ax - data[range_offs][0]) * (avg_y - point_ay)
                )
                * 0.5
            )

            if area > max_area:
                max_area = area
                max_area_point = data[range_offs]
                next_a = range_offs  # Next a is this b
            range_offs += 1

        sampled.append(max_area_point)  # Pick this point from the bucket
        a = next_a  # This a is the next a (chosen b)

    sampled.append(data[len(data) - 1])  # Always add last

    return sampled


def mult_by_scalar(
    vector: List[Tuple[int, float]], scalar: float
) -> List[Tuple[int, float]]:
    return [(x[0], x[1] * scalar) for x in vector]


def mean_every(data: List[Tuple[int, float]], seconds: int) -> List[Tuple[int, float]]:
    first_timestamp = data[0][0]
    last_timestamp = data[-1][0]
    new_data = []
    time_delta = last_timestamp - first_timestamp
    windows_count = math.ceil(time_delta / seconds)
    windows_length = len(data) // windows_count
    for window in itertools.batched(data, windows_length):
        mean = sum([w[1] for w in window]) / len(window)
        middle_index = len(window) // 2
        new_data.append((window[middle_index][0], mean))
    return new_data


def copy_labels(data_path: str, label_path: str, output_path: str):
    with open(data_path, "r") as data, open(label_path, "r") as label, open(
        output_path, "w"
    ) as output:
        data = data.readlines()
        label = label.readlines()
        data_headers = data[0].strip().split(",")
        label_headers = label[0].strip().split(",")

        assert len(data) == len(label)
        assert len(data) > 1
        assert len(data_headers) == len(label_headers)
        assert data_headers[0] == label_headers[0]
        assert data_headers[1] == label_headers[1]
        assert data_headers[2] == label_headers[2]

        wr = csv.writer(output, delimiter=",", quotechar='"', quoting=csv.QUOTE_MINIMAL)
        new_header = ["timestamp", "value", "is_anomaly"]
        wr.writerow(new_header)
        for data_line, label_line in zip(data[1:], label[1:]):
            parts = data_line.strip().split(",")
            timestamp, value = parts[0], parts[1]
            is_anomaly = label_line.strip().split(",")[2]
            new_line = [timestamp, value, is_anomaly]
            wr.writerow(new_line)


class ConversionMode(enum.Enum):
    W2L = "W2L"
    S2L = "S2L"
    CPY = "CPY"


def create_parser() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Convert between different CSV formats relating to time series analysis"
    )
    subparsers = parser.add_subparsers()

    w2l_parser = subparsers.add_parser(
        ConversionMode.W2L.value, help="Westermo to Labelstudio"
    )
    w2l_parser.set_defaults(mode=ConversionMode.W2L.value)
    w2l_parser.add_argument(
        "--data",
        "-d",
        help="Dataset input CSV file",
        type=str,
        required=True,
    )
    w2l_parser.add_argument(
        "--output",
        "-o",
        help="Output CSV path",
        type=str,
        required=True,
    )
    w2l_parser.add_argument(
        "--column",
        "-c",
        help="Column name to extract (default: 'value')",
        type=str,
        default="value",
    )
    w2l_parser.add_argument(
        "--threshold",
        "-t",
        help="Threshold for downsampling",
        type=int,
    )
    w2l_parser.add_argument(
        "--scale",
        "-s",
        help="Scale the data by a scalar",
        type=float,
    )
    w2l_parser.add_argument(
        "--mean",
        "-m",
        help="Take the mean of the data by <time>",
        type=str,
    )

    s2l_parser = subparsers.add_parser(
        ConversionMode.S2L.value, help="Label Studio to Label"
    )
    s2l_parser.set_defaults(mode=ConversionMode.S2L.value)
    s2l_parser.add_argument(
        "--data",
        "-d",
        help="Dataset input CSV file",
        type=str,
        required=True,
    )
    s2l_parser.add_argument(
        "--labels",
        "-l",
        help="Label input file",
        type=str,
        required=True,
    )
    s2l_parser.add_argument(
        "--output",
        "-o",
        help="Output CSV path",
        type=str,
        required=True,
    )
    s2l_parser.add_argument(
        "--column",
        "-c",
        help="Column name to extract",
        type=str,
        default="value",
    )
    s2l_parser.add_argument(
        "--labelid",
        "-i",
        help="Label Sudio label ID (to extract)",
        type=int,
        required=True,
    )

    cpy_parser = subparsers.add_parser(ConversionMode.CPY.value, help="Copy labels")
    cpy_parser.set_defaults(mode=ConversionMode.CPY.value)
    cpy_parser.add_argument(
        "--data",
        "-d",
        help="Dataset input CSV file",
        type=str,
        required=True,
    )
    cpy_parser.add_argument(
        "--source",
        "-s",
        help="Label input file",
        type=str,
        required=True,
    )
    cpy_parser.add_argument(
        "--output",
        "-o",
        help="Output CSV path",
        type=str,
        required=True,
    )

    return parser.parse_args()


def time_string_to_seconds(time_string: str) -> int:
    # Convert a time string to seconds
    # e.g. 1h -> 3600, 1d -> 86400
    time_dict = {"s": 1, "m": 60, "h": 3600, "d": 86400}
    time_unit = time_string[-1]
    time_value = int(time_string[:-1])
    return time_value * time_dict[time_unit]


# def westermo_to_labelstudio(
#     data: List[Tuple[int, float]],
#     output_path: str,
# ):
#     with open(output_path, "w") as output:
#         wr = csv.writer(output, delimiter=",", quotechar='"', quoting=csv.QUOTE_MINIMAL)
#         new_header = ["timestamp", "value"]

#         wr.writerow(new_header)
#         for timestamp, value in data:
#             # convert unix timestamp to human readable format
#             realtime = time.strftime(
#                 "%Y-%m-%d %H:%M:%S", time.localtime(int(timestamp))
#             )
#             new_line = [timestamp] + [value]
#             wr.writerow(new_line)


def westermo_to_label(
    data: List[Tuple[int, float]],
    output_path: str,
):
    with open(output_path, "w") as output:
        wr = csv.writer(output, delimiter=",", quotechar='"', quoting=csv.QUOTE_MINIMAL)
        new_header = ["timestamp", "value", "is_anomaly"]
        wr.writerow(new_header)
        for timestamp, value in data:
            # convert unix timestamp to human readable format
            realtime = time.strftime(
                "%Y-%m-%d %H:%M:%S", time.localtime(int(timestamp))
            )
            new_line = [timestamp] + [value] + ["0"]
            wr.writerow(new_line)


def overlaps(a: tuple, b: tuple) -> bool:
    a_start, a_end = a
    b_start, b_end = b
    # check if the two regions overlap anywhere
    return a_start < b_start < a_end or a_start < b_end < a_end


def index_in_regions(index: int, regions: list) -> bool:
    for region in regions:
        if region[0] <= index <= region[1]:
            return True
    return False


def label_studio_to_label(
    data: List[Tuple[int, float]],
    label_studio_path: str,
    output_path: str,
    label_id: int,
):
    with open(label_studio_path, "r") as label, open(output_path, "w") as output:
        # read the label studio json
        label_json = json.load(label)
        # extract the labels for the given label id
        label_json = list(filter(lambda x: x["id"] == label_id, label_json))
        if not label_json:
            print("Label ID not found")
            sys.exit(1)

        # extract the start and end of the labels
        labels = label_json[0]["label"]
        start_end = [(label["start"], label["end"]) for label in labels]

        # check if any labels overlap
        for _, label in enumerate(start_end):
            for _, other_label in enumerate(start_end):
                if label != other_label and overlaps(label, other_label):
                    print("Labels overlap")
                    sys.exit(1)

        wr = csv.writer(output, delimiter=",", quotechar='"', quoting=csv.QUOTE_MINIMAL)
        new_header = ["timestamp", "value", "is_anomaly"]

        wr.writerow(new_header)
        for timestamp, value in data:
            is_anomaly = 1 if index_in_regions(timestamp, start_end) else 0
            new_line = [timestamp] + [value] + [str(is_anomaly)]
            wr.writerow(new_line)


def read_data(args: argparse.Namespace) -> List[Tuple[int, float]]:
    with open(args.data, "r") as f:
        data = f.readlines()
        headers = data[0].strip().split(",")
        if args.column not in headers:
            print(f"Column {args.column} not found in input file")
            sys.exit(1)
        column_index = headers.index(args.column)

        # split on comma and extract the first column and the column to downsample
        splits = [x.strip().split(",") for x in data[1:]]

        data = [(int(x[0]), float(x[column_index])) for x in splits]

        if len(data[0]) != 2:
            print("Input file must have two columns")
            sys.exit(1)

        return data


if __name__ == "__main__":
    args = create_parser()

    if not hasattr(args, "mode"):
        print("No subcommand specified. Use -h for help")
        sys.exit(1)

    data: List[Tuple[int, float]] = []
    if args.column is not None:
        data = read_data(args)

    match args.mode:
        # case ConversionMode.W2S.value:
        #     westermo_to_labelstudio(data, args.output)
        case ConversionMode.W2L.value:
            if args.threshold is not None:
                data = largest_triangle_three_buckets(data, args.threshold)
            if args.scale is not None:
                data = mult_by_scalar(data, args.scale)
            if args.mean is not None:
                if (
                    args.mean[-1] not in ["s", "m", "h", "d"]
                    or not args.mean[:-1].isdigit()
                ):
                    print("Invalid time unit")
                    sys.exit(1)
                data = mean_every(data, time_string_to_seconds(args.mean))
            westermo_to_label(data, args.output)

        case ConversionMode.S2L.value:
            label_studio_to_label(data, args.labels, args.output, args.labelid)

        case ConversionMode.CPY.value:
            copy_labels(args.data, args.source, args.output)

        case _:
            print("Invalid mode")
