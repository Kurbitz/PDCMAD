# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app/nala
# Install python
RUN apt-get update
RUN apt-get install -y python3.11
RUN apt-get install -y python3-pip
RUN apt-get install -y python3-venv

# Copy anomaly detection source code
COPY nala/anomaly_detection/ ./anomaly_detection/
# Set up virtual environment and install python dependencies
RUN python3.11 -m venv anomaly_detection/venv
RUN . anomaly_detection/venv/bin/activate
ENV PATH="/app/nala/anomaly_detection/venv/bin:$PATH"
RUN python3.11 -m pip install -r anomaly_detection/requirements.txt

# Copy dependencies
WORKDIR /app
COPY nala/go.mod nala/go.sum ./nala/
COPY internal/ ./internal/

# Get Go dependencies
WORKDIR /app/nala
RUN go mod download

# Copy source code
COPY nala/*.go ./
COPY nala/testpy.py ./

# Build nala binary
RUN go build -o /nala

# Expose port 8088 to the outside world
EXPOSE 8088

# Command to run the executable, this is the entrypoint of the container
CMD [ "/nala" ]