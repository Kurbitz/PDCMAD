package main

import (
	"log"
	"os"
)

// The main entry point of the program
// All it does is call the Run function of the App specified in cli_setup.go
func main() {
	if err := App.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
