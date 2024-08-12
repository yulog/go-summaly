package main

import (
	"log/slog"
	"os"

	"github.com/yulog/go-summaly/server"
)

const name = "summaly"

const version = "0.0.2"

var revision = "HEAD"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slog.SetDefault(logger)
	server.New().SetVersion(version).Start()
}
