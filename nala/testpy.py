import sys
import pandas
from sklearn.ensemble import IsolationForest
from datetime import datetime
import numpy as np
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler


print(sys.argv[1]) # Prints argument on index 1 ie argument 2