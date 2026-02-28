package main

import (
	"log/slog"
)

func main() {
	// Good examples - these should pass
	slog.Info("starting server on port 8080")
	slog.Debug("database connection established")
	slog.Warn("cache miss for key")
	slog.Error("failed to connect to database")

	// Bad examples - these should fail

	// Violation: starts with uppercase
	slog.Info("Starting server")

	// Violation: contains non-English characters
	slog.Info("запуск сервера")

	// Violation: contains special characters
	slog.Info("server started!!!")
	slog.Error("connection!!! failed")

	// Violation: contains sensitive words
	slog.Info("user password test")
	slog.Debug("api_key value")

	password := "pass"
	api_key := "api"
	secretinfo := "secret"

	slog.Info("pass:" + password)
	slog.Info("api=" + api_key)
	slog.Info("secret" + secretinfo)

	slog.Info("secret info")

}
