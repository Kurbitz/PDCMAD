from pathlib import Path
from typing import Dict, Any

import numpy as np


from timeeval.algorithm import Algorithm
from timeeval.timeeval import TimeEval
from timeeval.data_types import TrainingType, InputDimensionality
from timeeval.adapters.function import FunctionAdapter
from timeeval.algorithms.subsequence_if import subsequence_if
from timeeval.params.base import FixedParameters
from timeeval.metrics.auc_metrics import PrAUC
from timeeval.params.bayesian import BayesianParameterSearch
from timeeval.params.grid_search import IndependentParameterGrid
from timeeval.params.grid_search import FullParameterGrid
from optuna.distributions import IntDistribution
from timeeval.integration.optuna.config import OptunaStudyConfiguration
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
from timeeval.algorithms.random_black_forest import random_black_forest
from timeeval.algorithms.hotsax import hotsax
from timeeval.algorithms.phasespace_svm import phasespace_svm
from timeeval.algorithms.deepant import deepant
from timeeval.algorithms.subsequence_if import subsequence_if
from timeeval.algorithms.iforest import iforest
from timeeval.algorithms.median_method import median_method
from timeeval.algorithms.generic_rf import generic_rf
from timeeval.algorithms.fft import fft
from timeeval.algorithms.arima import arima
from timeeval.algorithms.s_h_esd import s_h_esd


def threshold(
    data: AlgorithmParameter, args: Dict[str, Any]
) -> np.ndarray:
    if not isinstance(data, np.ndarray):
        raise ValueError("Data type not supported")
    scalar = args.get("scalar", 1.0)
    direction = args.get("direction", "greater")
    if direction == "greater":
        return np.where(data > scalar, 1.0, 0.0)
    elif direction == "less":
        return np.where(data < scalar, 1.0, 0.0)
    else:
        raise ValueError("Invalid direction")

def absolute_score(scores: np.ndarray, args: dict) -> np.ndarray:
    return np.where(scores / max(1,scores.max()) > 0.5,1.0,0.0)


def ones(
    data: AlgorithmParameter, args: Dict[str, Any]
) -> np.ndarray:
    if not isinstance(data, np.ndarray):
        raise ValueError("Data type not supported")
    ones = np.ones(data.shape)
    ones[0:10] = 0
    return ones


def main():
    dm = DatasetManager(Path("test/w_data/"), create_if_missing=False)
    thermal_datasets = dm.select(collection="sys-thermal")
    cpu_datasets = dm.select(collection="cpu-user")
    smb_datasets = dm.select(collection="sys-mem-buffered")

    cpu_algorithms = [
        # dwt_mlead(),
        # kmeans(params=FixedParameters({"anomaly_window_size": 200, "n_clusters": 2})),
        # grammarviz3(params=FixedParameters({"alphabet_size": 4, "anomaly_window_size": 200, "paa_transform_size": 3})),
        # subsequence_lof(params=FixedParameters({"n_neighbors": 50, "window_size": 60})),
        # subsequence_if(params=FixedParameters({"window_size": 75})),
        # lstm_ad(params=FixedParameters({"window_size": 27})),
        Algorithm(
            name="ReferenceAlgorithm",
            main=FunctionAdapter(threshold),
            data_as_file=False,
            param_config=FixedParameters(
                {
                    "scalar": 0.625,
                }
            ),
        ),
        

    ]    

    thermal_algo = [
        # dwt_mlead(),
        # kmeans(params=FixedParameters({"anomaly_window_size": 70, "n_clusters": 120})),
        # grammarviz3(params=FixedParameters({"alphabet_size": 2, "anomaly_window_size": 200, "paa_transform_size": 3})),
        # subsequence_lof(params=FixedParameters({"window_size": 25, "n_neighbors": 800})),
        # subsequence_if(params=FixedParameters({"window_size": 25, "n_trees": 150})),
        # lstm_ad(params=FixedParameters({"window_size": 25})),
        Algorithm(
            name="ReferenceAlgorithm",
            main=FunctionAdapter(threshold),
            data_as_file=False,
            param_config=FixedParameters(
                {
                    "scalar": 1.68,
                }
            ),
        ),
    ]

    smb_algorithms = [
        # dwt_mlead(),
        # kmeans(params=FixedParameters({"anomaly_window_size": 320, "n_clusters": 2})),
        # grammarviz3(params=FixedParameters({"alphabet_size": 4, "anomaly_window_size": 100, "paa_transform_size": 3})),
        # subsequence_lof(params=FixedParameters({"n_neighbors": 512, "window_size": 4})),
        # subsequence_if(params=FixedParameters({"n_trees": 32, "window_size": 64})),
        # lstm_ad(params=FixedParameters({"window_size": 8})),
        Algorithm(
            name="ReferenceAlgorithm",
            main=FunctionAdapter(threshold),
            data_as_file=False,
            param_config=FixedParameters(
                {
                    "scalar": 1215494622,
                    "direction": "less",
                }
            ),
        ),
    ]

    thermal_timeval = TimeEval(
        dm,
        thermal_datasets,
        thermal_algo,
        metrics=[
            DefaultMetrics.ROC_AUC,
            DefaultMetrics.RANGE_PR_AUC,
            DefaultMetrics.AVERAGE_PRECISION,
            DefaultMetrics.PR_AUC,
        ],
        results_path=TimeEval.DEFAULT_RESULT_PATH.joinpath("thermal"),
    )
    cpu_timeval = TimeEval(
        dm,
        cpu_datasets,
        cpu_algorithms,
        metrics=[
            DefaultMetrics.ROC_AUC,
            DefaultMetrics.RANGE_PR_AUC,
            DefaultMetrics.AVERAGE_PRECISION,
            DefaultMetrics.PR_AUC,
        ],
        results_path=TimeEval.DEFAULT_RESULT_PATH.joinpath("cpu"),
    )

    smb_timeval = TimeEval(
        dm,
        smb_datasets,
        smb_algorithms,
        metrics=[
            DefaultMetrics.ROC_AUC,
            DefaultMetrics.RANGE_PR_AUC,
            DefaultMetrics.AVERAGE_PRECISION,
            DefaultMetrics.PR_AUC,
        ],
        results_path=TimeEval.DEFAULT_RESULT_PATH.joinpath("smb"),
    )

    thermal_timeval.run()
    results = thermal_timeval.get_results(aggregated=False)
    print(results)
    cpu_timeval.run()
    results = cpu_timeval.get_results(aggregated=False)
    print(results)
    smb_timeval.run()
    results = smb_timeval.get_results(aggregated=False)
    print(results)


if __name__ == "__main__":
    main()
