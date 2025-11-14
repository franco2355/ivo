package main

import (
	"proyecto-integrador/app"
	"proyecto-integrador/db"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Info("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	db.StartDbEngine()
	app.StartRoute()
}
