import sys
import pandas as pd

# import matplotlib.pyplot as plt
from sklearn.ensemble import IsolationForest
from datetime import datetime

# import plotly.express as px
import numpy as np
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler


def preprocess_data(file_path):
    # Read CSV file
    df = pd.read_csv(file_path)

    # lets assume data starts coming from this time
    start_time = datetime(2023, 10, 29, 0, 0, 0)
    time_range = pd.date_range(start=start_time, periods=len(df), freq="30S")
    # adding time to data after every 30 from that date
    df["timestamp"] = time_range

    return df


def apply_PCA(df, keep_var=0.95):  # Keep 95% of variance
    # Scaling the data to make it more normally distributed before applying PCA
    scaler = StandardScaler()
    df_scaled = scaler.fit_transform(df)

    # Applying PCA
    pca_model = PCA(n_components=keep_var)
    df_pca = pca_model.fit_transform(df_scaled)

    return df_pca, pca_model


def apply_IF(df_pca, df, no_of_tree=1000, perchentage_of_outlier=0.01):
    iso_forest = IsolationForest(
        n_estimators=no_of_tree, contamination=perchentage_of_outlier
    )
    iso_forest.fit(df_pca)

    # Prediction
    anomalies = iso_forest.predict(df_pca)
    anomaly_indices = np.where(anomalies == -1)[0]

    # output dataframe with anomaly indices(log file)
    df_anomaly = df.copy()
    df_anomaly["outliers"] = "no"
    df_anomaly.loc[df_anomaly.index[anomaly_indices], "outliers"] = "yes"

    # # output dataframe with anomaly indices(log file)
    # df_anomalies = df.iloc[anomaly_indices]

    return df_anomaly


def main(args):
    try:
        csv_input_path = args[1]
        csv_output_path = args[2]
    except IndexError:
        raise SystemExit(f"Usage: {args[0]}, inputfile, outputfile")
    prep_df = preprocess_data(file_path=csv_input_path)

    df_pca, _ = apply_PCA(
        prep_df.drop("timestamp", axis=1)
    )  # didn't take timestamp as feature

    output_df = apply_IF(df_pca, prep_df)

    anomaly_log = output_df[output_df["outliers"] == "yes"]

    # Write the results to a new CSV file
    anomaly_log.to_csv(csv_output_path, index=False)


if __name__ == "__main__":
    main(sys.argv)
