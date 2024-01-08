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


### Usage

* How to run the program
* Step-by-step bullets
```
code blocks for commands
```

## Environment Variables

## Help

Any advise for common problems or issues.
```
command to run if program contains helper info
```

## Authors

Contributors names and contact info

ex. Dominique Pizzie  
ex. [@DomPizzie](https://twitter.com/dompizzie)