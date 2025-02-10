package main

import (
	"bbrew/internal/services"
	"log"
)

func main() {
	appService := services.NewAppService()
	if err := appService.Boot(); err != nil {
		log.Fatalf("Error initializing data: %v", err)
	}
	appService.BuildApp()

	if err := appService.GetApp().Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
