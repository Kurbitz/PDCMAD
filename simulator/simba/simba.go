package main

import (
	"log"
	"os"
	simba "pdc-mad/simba/internal"
)

func main() {

	if err := simba.App.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
