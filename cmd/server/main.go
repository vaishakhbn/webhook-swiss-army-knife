package main

import (
	"log/slog"
	"os"

	"github.com/vaishakhbn/webhook-swiss-army-knife/internal/httpserver"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := httpserver.Run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
