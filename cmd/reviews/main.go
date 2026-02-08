package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load() // Загружает переменные из .env файла
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbUser := os.Getenv("POSTGRES_USER")
	fmt.Println(dbUser)
}
