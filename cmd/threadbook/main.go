package main

import (
	"log"

	"github.com/onionfriend2004/threadbook_backend/internal/app"
)

func main() {
	log.Printf("Hello, world!")

	err := app.Run()

	if err != nil {
		log.Printf("server is not running")
		return
	}
}
