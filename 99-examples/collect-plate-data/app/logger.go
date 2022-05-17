package app

import (
	"log"
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	var err error
	if os.Getenv("ENV") == "production" {
		Logger, err = zap.NewProduction()
	} else {
		Logger, err = zap.NewDevelopment()
	}
	if err != nil {
		log.Fatal(err)
	}
}
