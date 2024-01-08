# Nala

Nala is a RestAPI that interfaces between the database and the anomaly detection algorithm.

## Description

This was created as part of a school project for Software Engineering 2 at Mälardalen University.
This tool handles the interaction between user, database, and anomaly detection algorithms within PDC-MAD using parametric endpoints. 

## Getting Started

### Dependencies

* Python 3.11
* Go 1.20

## Usage
### Anomaly detection
```shell
curl localhost:8088/[Algorithm]/[host]/[duration]
```
This will trigger given detection algorithm using data from the host of given duration. 
It is done by querying the data from the database with given parameters. 
#### Example
```shell
curl localhost:8088/IF/system-1/36h
```
This will trigger Isolation Forest on data from system-1 taking data from 36 hours ago to now.

### Status
```shell
curl localhost:8088/status
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
