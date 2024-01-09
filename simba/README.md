# Simba
Simulation with bad actors.
### First time setup

If you have not set up InfluxDB yet then use this command.
```shell
docker run -p 8086:8086 --network-alias influxdb -v myInfluxVolume:/var/lib/influxdb2 influxdb:latest
```

Go to localhost:8086 and follow the tutorial. Create a bucket called "metrics". Save the token generated and the organization name as it will be needed for Simba. When this is done you can start using Simba.

When current directory is inside simulator/simba:
```shell
go run . --dbtoken [GENERATED TOKEN] PATH/TO/CSV/FILE
```
This will then fill the database with the data provided. You might run into errors like `Write error: not found: bucket "metrics" not found` or `Write error: not found: organization name "test" not found`. 

In the case of bucket not being find: Create a bucket on your local influxdb named "metrics". 

In case of organization not being found: input the organization name on line 25 in "influxdbapi.go".

If everything is done correctly then you should have data in your local InfluxDB under the "Data Explorer" tab.

#### Common arguments
* start-at: how far into the file given Simba should start the reading.
* duration: from the starting point how much of the data should be written in time string eg. 5d for 5 days worth of data.
* anomaly: given predefined anomaly injection algorithm, inject anomalous data instead of raw data.
#### Fill
Filling the database from given data or injected anomalies. The data simulated can be specified with the common arguments as well as:
* gap: how much time should be left between the last inserted datapoint and now.

#### Stream
Simulate a running server by sending data every 30 seconds. Besides the common arguments there are:
* time-multiplier: an int which multiplies the base 30s time interval to run `n` times faster.
* append: if you want to simulate from the latest datapoint of given system. If this is used the given system needs have data in the database already.