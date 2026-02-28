package main

import (
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()

	// Violations:
	logger.Info("Starting server")   // uppercase
	logger.Info("запуск сервера")    // non-English
	logger.Info("server started!")   // special chars

	password := "secret"
	logger.Info("password: " + password) // sensitive variable
}
