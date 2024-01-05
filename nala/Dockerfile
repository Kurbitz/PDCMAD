# syntax=docker/dockerfile:1

FROM golang:1.20

WORKDIR /app

COPY nala/go.mod nala/go.sum ./nala/
COPY internal/ ./internal/

WORKDIR /app/nala
RUN go mod download

COPY nala/*.go ./
COPY nala/testpy.py ./

RUN go build -o /nala

RUN apt-get update
RUN apt-get install -y python3.11
RUN apt-get install -y python3-pip
RUN apt-get install -y python3-venv

COPY nala/anomaly_detection/ ./anomaly_detection/
RUN python3.11 -m venv anomaly_detection/venv
RUN . anomaly_detection/venv/bin/activate
ENV PATH="/app/nala/anomaly_detection/venv/bin:$PATH"
RUN python3.11 -m pip install -r anomaly_detection/requirements.txt

EXPOSE 8088

CMD [ "/nala" ]