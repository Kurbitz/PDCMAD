package main

import (
	"internal/logger"
	"log"
	"os"
)

func main() {
	logger.NewLogger()
	if err := App.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
