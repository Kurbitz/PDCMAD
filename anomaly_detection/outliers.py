import sys
import pandas as pd

# import matplotlib.pyplot as plt
from sklearn.ensemble import IsolationForest
import time
from sklearn.preprocessing import StandardScaler


def read_file(file_path):
    # Read CSV file
    df = pd.read_csv(file_path)

    return df


def train_prediction_IF(
    dataframe, feature, no_of_tree=1000, perchentage_of_outlier=0.01
):
    df = pd.DataFrame({"server-up": dataframe["server-up"], "feature": feature})

    scaler = StandardScaler()
    df_scaled = scaler.fit_transform(df)

    model = IsolationForest(
        n_estimators=no_of_tree, contamination=perchentage_of_outlier
    )
    prediction = model.fit_predict(df_scaled[:])

    return prediction


def run_train_predction_IF_for_every_column(input_df):
    start_time = time.time()

    anomalies = []  # empty list
    # running IF with  two(any_column, server-up) column, not taking timestamp
    for col in input_df.columns[0:]:
        if col != "server-up" and col != "timestamp":
            anomaly = train_prediction_IF(dataframe=input_df, feature=input_df[col])
            anomalies.append(anomaly)

    end_time = time.time()

    # calculating time
    elapsed_time = end_time - start_time
    print("Elapsed time(min): ", elapsed_time / 60)

    return anomalies


def sum_all_annomaly_in_outpuDF(input_df, anomalies_2D):
    # Annomalies are like 1(regular) or -1(anomaly), replacing it like 0(regular) or 1(anomaly)
    anomalies_2D = [
        [1 if item == -1 else 0 for item in sublist] for sublist in anomalies_2D
    ]

    df_anomaly = pd.DataFrame(columns=input_df.columns)
    df_anomaly[["timestamp", "server-up"]] = input_df[["timestamp", "server-up"]].copy()
    anomaly_index = 0
    for column in df_anomaly.columns:
        if column not in ["timestamp", "server-up"]:
            df_anomaly[column] = anomalies_2D[anomaly_index]
            anomaly_index += 1

    return df_anomaly


# def apply_PCA(df, keep_var=0.95):  # Keep 95% of variance
#     # Scaling the data to make it more normally distributed before applying PCA
#     scaler = StandardScaler()
#     df_scaled = scaler.fit_transform(df)

#     # Applying PCA
#     pca_model = PCA(n_components=keep_var)
#     df_pca = pca_model.fit_transform(df_scaled)

#     return df_pca, pca_model


# def apply_IF(df_pca, df, no_of_tree=1000, perchentage_of_outlier=0.01):
#     iso_forest = IsolationForest(
#         n_estimators=no_of_tree, contamination=perchentage_of_outlier
#     )

#     # Prediction
#     anomalies = iso_forest.fit_predict(df_pca)
#     anomaly_indices = np.where(anomalies == -1)[0]

#     # output dataframe with anomaly indices(log file)
#     df_anomaly = df.copy()
#     df_anomaly["outliers"] = "no"
#     df_anomaly.loc[df_anomaly.index[anomaly_indices], "outliers"] = "yes"

#     # # output dataframe with anomaly indices(log file)
#     # df_anomalies = df.iloc[anomaly_indices]

#     return df_anomaly


def main(args):
    try:
        csv_input_path = args[1]
        csv_output_path = args[2]
    except IndexError:
        raise SystemExit(f"Usage: {args[0]}, inputfile, outputfile")

    input_df = read_file(file_path=csv_input_path)

    anomalies_2d = run_train_predction_IF_for_every_column(input_df=input_df)
    output_df = sum_all_annomaly_in_outpuDF(input_df, anomalies_2d)

    # Write the results to a new CSV file
    output_df.to_csv(csv_output_path, index=False)


if __name__ == "__main__":
    main(sys.argv)
