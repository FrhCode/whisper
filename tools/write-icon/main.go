package main

import (
	"log"
	"os"

	"whispr/internal/icon"
)

func main() {
	if err := os.WriteFile("assets/whispr.ico", icon.Data(), 0644); err != nil {
		log.Fatal(err)
	}
}
