from pathlib import Path
from typing import Dict, Any

import numpy as np


from timeeval.algorithm import Algorithm
from timeeval.timeeval import TimeEval
from timeeval.data_types import TrainingType, InputDimensionality
from timeeval.adapters.function import FunctionAdapter
from timeeval.algorithms.subsequence_if import subsequence_if
from timeeval.params.base import FixedParameters
from timeeval.metrics import DefaultMetrics
from timeeval.adapters.function import FunctionAdapter
from timeeval.data_types import AlgorithmParameter
from timeeval.datasets.dataset_manager import DatasetManager
from timeeval.algorithms.cof import cof
from timeeval.algorithms.subsequence_lof import subsequence_lof
from timeeval.algorithms.kmeans import kmeans
from timeeval.algorithms.torsk import torsk
from timeeval.algorithms.lstm_ad import lstm_ad
from timeeval.algorithms.grammarviz3 import grammarviz3
from timeeval.algorithms.dwt_mlead import dwt_mlead
from timeeval.algorithms.donut import donut
from timeeval.algorithms.normalizing_flows import normalizing_flows


def your_algorithm_function(
    data: AlgorithmParameter, args: Dict[str, Any]
) -> np.ndarray:
    if not isinstance(data, np.ndarray):
        raise ValueError("Data type not supported")
    scalar = args.get("scalar", 1.0)
    return np.where(data > scalar, 1.0, 0.0)


def main():
    dm = DatasetManager(Path("test/w_data/"), create_if_missing=False)
    datasets = dm.select()
    algorithms = [
        # donut(),
        # lstm_ad(),
        # torsk(),
        # grammarviz3(),
        # cof(params=FixedParameters({"n_neighbors": 20, "random_state": 42})),
        Algorithm(
            name="MyPythonFunctionAlgorithm",
            main=FunctionAdapter(your_algorithm_function),
            data_as_file=False,
            param_config=FixedParameters(
                {
                    "scalar": 0.95,
                }
            ),
        ),
        Algorithm(
            name="MyPythonFunctionAlgorithm",
            main=FunctionAdapter(your_algorithm_function),
            data_as_file=False,
            param_config=FixedParameters(
                {
                    "scalar": 1.5,
                }
            ),
        ),
        subsequence_lof(params=FixedParameters({"window_size": 50, "n_neighbors": 40})),
        kmeans(params=FixedParameters({"anomaly_window_size": 200, "n_clusters": 40})),
        kmeans(),
        # dwt_mlead(),
        # list of algorithms which will be executed on the selected dataset(s)
        # calling customized algorithm
    ]
    timeeval = TimeEval(
        dm,
        datasets,
        algorithms,
        metrics=[
            DefaultMetrics.ROC_AUC,
            DefaultMetrics.RANGE_PR_AUC,
            DefaultMetrics.AVERAGE_PRECISION,
            DefaultMetrics.PR_AUC,
        ],
    )
    timeeval.run()
    results = timeeval.get_results(aggregated=False)
    print(results)


if __name__ == "__main__":
    main()
