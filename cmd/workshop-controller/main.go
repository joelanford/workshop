package main

import (
	"os"

	log "github.com/go-kit/kit/log"
	"github.com/joelanford/workshop/cmd/workshop-controller/app"
)

func main() {
	logger := log.With(log.NewSyncLogger(log.NewLogfmtLogger(os.Stderr)),
		"ts", log.DefaultTimestampUTC,
		"component", "workshop-controller")

	if err := app.Run(logger); err != nil {
		logger.Log("msg", "application error", "err", err)
	}
}
