module pdc-mad/simba

go 1.20

require (
	github.com/gocarina/gocsv v0.0.0-20231116093920-b87c2d0e983a // indirect
	github.com/influxdata/influxdb-client-go/v2 v2.13.0 // indirect
	github.com/urfave/cli/v2 v2.26.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/oapi-codegen/runtime v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/exp v0.0.0-20240103183307-be819d1f06fc // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
)

replace internal/system_metrics => ../internal/system_metrics

replace internal/influxdbapi => ../internal/influxdbapi

require internal/system_metrics v1.0.0

require (
	github.com/schollz/progressbar/v3 v3.14.1
	internal/influxdbapi v1.0.0
)
