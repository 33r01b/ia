package main

import (
	"os"

	"github.com/33r01b/ia/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args))
}
