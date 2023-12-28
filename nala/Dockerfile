# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY anomaly_detection ./anomaly_detection

RUN apt-get update
RUN apt-get install -y python3.11
RUN apt-get install -y python3-pip
RUN apt-get install -y python3-venv

RUN python3.11 -m venv ./anomaly_detection/venv
RUN . ./anomaly_detection/venv/bin/activate
ENV PATH="/app/anomaly_detection/venv/bin:$PATH"

RUN python3.11 -m pip install -r anomaly_detection/requirements.txt

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY testpy.py ./

RUN go build -o /nala

EXPOSE 8088

CMD [ "/nala" ]