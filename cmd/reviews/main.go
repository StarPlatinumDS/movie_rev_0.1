package main

import (
	"log"
	"os"
)

func main() {
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}

	log.Println("Application shutdown successfull")
}
