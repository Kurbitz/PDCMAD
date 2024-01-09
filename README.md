#  PDC-MAD - Performance Data Collection, Monitoring and Anomaly Detection
![PDC-MAD octopus](PDCMAD-1600x1600.png)
## Description

PDC-MAD is a project that will help future development of anomaly detection algorithms for server data by simulating "normal" data with the ability to inject anomalous data. With the built in anomaly detection module you can trigger any self-defined anomaly detection algorithms. Both simulated and anomaly data are stored in [InfluxDB](#services) and visualized as graphs with [Grafana](#services).

It can either simulate in real-time or in batches. 

## Getting Started

### Dependencies

- [Docker](https://docs.docker.com/get-docker/): In order to run the docker stack.
- [Go 1.20](https://go.dev/doc/install): In order to run Simba and Nala as a standalone.

### Installing
This program is working on Windows 10 and 11 but the documentation will assume that you are a Linux user.
#### Pre-requisites
Clone PDC-MAD repository to desired location. Before deploying the containers you need to define the environment variables for the docker stack by creating a `.env` file. There is a `.env_example` in the docker folder so you know the general structure. They are further explained in the [Environment Variables](#environment-variables) section.

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

Grafana and InfluxDB is checked by going to a browser and see if the service is up.

Default InfluxDB:
`http://localhost:8086`

Default Grafana:
`http://localhost:3000`

Testing Simba is by running the installed `simba` command and getting help text.

Nala has a test command.
```shell
curl localhost:8088/test
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
Simulate 8h of data and start real-time simulation with anomalies
```shell
simba fill --duration 8h foo.csv;
simba stream --start-at 8h --append --anomalies cpu-user-high foo.csv
```
Simulate ten days of data taken from multiple datasets
```shell
simba fill --duration 2d --gap 10d foo1.csv;
simba fill --start-at 2d --duration 1d --gap 3d --anomaly cpu-user-sin foo2.csv;
simba fill --start-at 2d --duration 2d foo3.csv
```

### Nala
Nala is a RestAPI with a few built in commands. This could be hooked up to a web page or triggered by Bash scripts.

DISCLAIMER: All examples will assume the default settings of the container and API.
#### Trigger
General structure:
```shell
curl HOSTNAME:PORT/[ALGORITHM]/[SYSTEM-NAME]/[DURATION]
```
Trigger **I**solation **F**orest on `foo` with `36h` of data
```shell
curl localhost:8088/IF/foo/36h
```
#### Status
Check on status of anomaly detection. If it is done or not. This response can be used for automated anomaly detection.
```shell
curl localhost:8088/status
```
#### Test
Run a test to see if you can reach Nala.
```shell
curl localhost:8088/test
```

## Environment Variables
### Docker
List of variables .env
```
INFLUXDB_ADMIN_USER
INFLUXDB_ADMIN_PASSWORD
INFLUXDB_ORG
INFLUXDB_BUCKET
INFLUXDB_ADMIN_TOKEN
GRAFANA_ADMIN_USER
GRAFANA_ADMIN_PASSWORD
INFLUXDB_PORT
GRAFANA_PORT
NALA_PORT
```
- `INFLUXDB_ADMIN_USER` Admin username for InfluxDB.
- `INFLUXDB_ADMIN_PASSWORD` Admin password for InfluxDB.
- `INFLUXDB_ORG` InfluxDB predefined organization.
- `INFLUXDB_BUCKET` InfluxDB predefined bucket. This is where Simba will write and clean by default. Nala queries this bucket when getting the server data.
- `INFLUXDB_ADMIN_TOKEN` Token used for admins. This is what Nala uses when querying and writing data.
- `GRAFANA_ADMIN_USER` Admin username for Grafana
- `GRAFANA_ADMIN_PASSWORD` Admin password for Grafana
- `INFLUXDB_PORT` Default value of exposed port
- `GRAFANA_PORT` Default value of exposed port
- `NALA_PORT` Default value of exposed port

### Simba
Simba uses either a token provided with `--db-token` or an exported environment variable named `INFLUXDB_TOKEN`. These can be created in InfluxDB for each user separately.

## Help
Simba has help arguments (`-h`) for each subcommand.
See [Nala's](#nala) documentation for general structure.
Both [InfluxDB](https://docs.influxdata.com) and [Grafana](https://grafana.com/docs/) have their own well-defined documentation.

## Dataset
For both Nala and Simba we use Westermo's [test-system-performance-dataset](https://github.com/westermo/test-system-performance-dataset) data structure. Follow the link to get more info on all the metrics.


## Authors
- @Meeptard
- @Kurbitz
- @Ni7070
- @segerstrom
- @FerDovah
- @cenza1
