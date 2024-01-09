# First time setup
### Disclaimer
Some of the following commands will take into account that you are in specific work directories to actually work as intended. Please follow every step unless you know what you are doing!
## Prerequisites

You will need Python version 3.11

To find out which version that is your systems default use:
```shell
python --version
```
To find out if you have the correct version installed use:

*Linux*
```shell
which python3.11
```
*Windows*
```shell
where python3.11
```

### Creating the environment
**When you have checked and/or installed the correct python version for this project it is time for the first time setup.**

Run this command to setup the environment used in this project

**Note: this can be done with other environment modules but for highest reproducability, use the one listed below**
```shell
python3.11 -m venv PATH/TO/WORK/DIRECTORY
```
The path in this case will be "anomaly_detection", more importantly the directory with the "requirements.txt" file


When you have created the environment you change current directory to the environments root directory which is stated on the line above
```shell
cd PATH/TO/WORK/DIRECTORY
```

### Activating the environment
The next step is ***very important*** if you skip this step you will install every library to your global python which is usually not desired

When the environment is create there will be a new folder called **bin** can be checked with `ls`. Inside the bin folder there is a few scipts called **activate** with different file endings.

Depending on your current operating system you will run the file that corresponds to your system.

<h3>Windows</h3>

*cmd*
```shell
bin\activate.bat
```
*PowerShell*
```shell
bin\Scripts\Activate.ps1
```
<h3>Linux</h3>

*bash/zsh*
```shell
source myvenv/bin/activate
```
If you have any other shell in linux you are on your own.

### Download and install dependencies
This is the step that will install to global environment if your actvation of the environment went wrong somehow. Make sure that you have done the activation step correctly.

*Windows/Linux*
```shell
pip install -r requirements.txt
```

After all the installations are done you are ready to start developing, good luck!