import pandas as pd

INPUTFILE = "../nala/logs/go_output.csv"
OUTPUTFILE = "../nala/logs/py_output.csv"


# Just a bunch of prints
def main():
    dataframe = pd.read_csv(INPUTFILE)
    dataframe.to_csv(OUTPUTFILE)


main()
