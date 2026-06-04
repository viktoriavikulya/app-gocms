package main

import (
	"log"

	"github.com/fastygo/app-gocms/pkg/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
