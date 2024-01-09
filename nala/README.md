# Nala

Nala is a RestAPI that interfaces between the database and the anomaly detection algorithm.

## Description

This was created as part of a school project for Software Engineering 2 at Mälardalen University.
This tool handles the interaction between user, database, and anomaly detection algorithms within PDC-MAD using parametric endpoints. 

## Getting Started

### Dependencies

* Python 3.11
* Go 1.20
* Docker

### Installing

Clone PDC-MAD repository to desired location.

Go to the now downloaded folder.
```shell
cd /PATH/TO/PDC-MAD
```
You need to set up the .env file before building the container. there is an .env_example in the docker folder so you know what the structure looks like.

Next step is to build nala.
```shell
docker build --tag nala -f nala/Dockerfile .
```
This will take a few minutes depending on hardware.
### Setting up the docker stack

When it is done you can run it by going to the docker folder of PDC-MAD:
```shell
cd /PATH/TO/PDC-MAD
```
And running the container
```shell
docker compose up
```
This will set everything up.

### Parameters
There are a few parameters that the endpoint can handle.
**Test**
```shell
curl localhost:8088/nala/test
```
This will run a "smoketest" to test if the python environment is working.


**Trigger algorithm**
```shell
curl localhost:8088/nala/[Algorithm]/[host]/[duration]
```
This will trigger given algorithm using data from the host for the duration.
**Example**
```shell
curl localhost:8088/nala/IF/system-1/36h
```
This will trigger Isolation Forest on data from system-1 taking data from 36 hours ago to now.

**Status**
```shell
curl localhost:8088/nala/status
```
A status check if an anomaly detection algorithm is currently running 

## Help

If you get a missing parameter error when trying to start the container it probably means the environment variables in the docker have not been set up correctly.
How we have solved the error.
```shell
docker compose down
```
Remove the misconfigured files. Some might be protected.
```shell
rm -r influxdb/
rm -r grafana/data
```
And start the container again.
```shell
docker compose up
```

## Authors

@Meeptard
@Kurbitz
@Ni7070
@segerstrom
@FerDovah
@cenza1
