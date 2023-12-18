module pdc-mad/influxdbapi

go 1.20

replace internal/system_metrics => ../system_metrics

require (
	github.com/influxdata/influxdb-client-go/v2 v2.13.0
	internal/system_metrics v0.0.0-00010101000000-000000000000
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/gocarina/gocsv v0.0.0-20231116093920-b87c2d0e983a // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/oapi-codegen/runtime v1.0.0 // indirect
	golang.org/x/net v0.17.0 // indirect
)
