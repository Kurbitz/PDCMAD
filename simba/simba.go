package main

import (
	"log"
	"os"
)

func main() {
	if err := App.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
