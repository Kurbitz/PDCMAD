#  PDC-MAD - Performance Data Collection, Monitoring and Anomaly Detection
![PDC-MAD octopus](PDCMAD-1600x1600.png)
## Description

An in-depth paragraph about your project and overview of use.

## Getting Started

### Dependencies

* Describe any prerequisites, libraries, OS version, etc., needed before installing program.
* ex. Windows 10

### Installing

#### Pre-requisites
Clone PDC-MAD repository to desired location. Go to the repository root. Before deploying the containers you need to define the environment variables for the docker stack. There is a .env_example in the docker folder so you know the general structure. They are further explained in the [Environment Variables](#environment-variables) section.

#### Docker stack
First step is to build nala. This needs to be done from the repository root folder because of shared dependencies.
```shell
docker build --tag nala -f nala/Dockerfile .
```
This will take a few minutes depending on hardware. When it is done you can run it while in the docker folder.
```shell
docker compose up
```

#### Simba
You can either run or install Simba, we would recommend installing it. While in the simba folder.
```shell
go install .
```
In case of issues when installing the go program please read Go's [documentation](https://go.dev/doc/tutorial/compile-install).

#### Checking installation
Now everything should be set up. You can try if everything is properly installed and running with the checks below.

Grafana and InfluxDB is checked by going to browser and see if the service is up.

Default InfluxDB:
`http://localhost:8086`

Default Grafana:
`http://localhost:3000`

Testing Simba is by running the installed `simba` command and getting help text.

Nala has a status command which tells you if there is any algorithms running.
```shell
curl localhost:8088/status
```

If all these commands works you have successfully installed PDC-MAD!


## Services
#### InfluxDB
Accessed with (by default) `localhost:8086` you can log in and administer users, buckets and also test queries when developing further.
#### Grafana
The docker stack comes with preset datapoints and dashboards that you can use or otherwise set up your own inside of Grafana. Access to the service is by default `localhost:3000`.
## Usage
### Simba
There are four base subcommands in Simba. These can be combined and used in different ways to fulfill more complex functionality.

#### Fill
With a whole file
```shell
simba fill foo.csv
```
Specific duration
```shell
simba fill --duration 5d foo.csv
```
Specific part
```shell
simba fill --start-at 3d --duration 1d foo.csv
```
Leave room for more data
```shell
simba fill --duration 5d --gap 3d foo.csv
```
Inject anomalous data
```shell
simba fill --duration 1h --anomaly cpu-user-high foo.csv
```
#### Stream
Start a real-time server simulation
```shell
simba stream foo.csv
```
Append real-time data to foo
```shell
simba stream --append foo.csv
```
Inject anomalies real-time
```shell
simba stream --anomaly cpu-user-sin foo.csv
```
Specific part
```shell
simba stream --start-at 3d --duration 25m foo.csv
```
Simulate at 20 times the speed
```shell
simba stream --time-multiplier 20 foo.csv
```
#### Clean
Remove data for all host on specified [INFLUXDB_BUCKET](#environment-variables)
```shell
simba clean --all
```
Remove last 1h of data
```shell
simba clean --duration 1h
```
Remove whole measurements
```shell
simba clean --all -M anomalies
```
#### Combinations
As the append argument is not available for Fill you have to shift the data with gap and calculate the next starting point and gap.

Simulate five days with one day of anomalous data
```shell
simba fill --duration 2d --gap 5d foo.csv;
simba fill --start-at 2d --duration 1d --gap 3d --anomaly cpu-user-sin foo.csv;
simba fill --start-at 2d --duration 2d foo.csv
```
Simulate three days of data and start real-time simulation
```shell
simba fill --duration 3d foo.csv;
simba stream --start-at 3d --append foo.csv
```
Simulate ten days of data taken from multiple datasets
```shell
simba fill --duration 2d --gap 10d foo1.csv;
simba fill --start-at 2d --duration 1d --gap 3d --anomaly cpu-user-sin foo2.csv;
simba fill --start-at 2d --duration 2d foo3.csv
```

### Nala

## Environment Variables

## Help

Any advise for common problems or issues.
```
command to run if program contains helper info
```

## Dataset

## Authors
