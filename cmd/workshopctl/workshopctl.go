package main

import (
	"fmt"
	"os"

	"github.com/joelanford/workshop/cmd/workshopctl/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}
}
