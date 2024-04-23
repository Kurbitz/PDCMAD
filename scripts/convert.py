import time
import sys
import csv


def main(filename: str, column: str, category: str):
	with open(filename, "r") as input, open("output.csv", "w") as output:
		lines = input.readlines()
		wr = csv.writer(output, delimiter=",", quotechar='"', quoting=csv.QUOTE_MINIMAL)
		
		header = lines[0].split(",")
		header[-1] = header[-1].strip()
		column_index = header.index(column)
		new_header = ["date", "category", "value"]

		wr.writerow(new_header)
		for line in lines[1:]:
			cols = line.split(",")
			timestamp = cols[0]
			# convert unix timestamp to human readable format
			realtime = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(int(timestamp)))
			cols[-1] = cols[-1].strip()
			new_line = [realtime] + [category] + [cols[column_index]]
			wr.writerow(new_line)


if __name__ == "__main__":
	filename = sys.argv[1]
	column = sys.argv[2]
	category = sys.argv[3]
	main(filename, column, category)