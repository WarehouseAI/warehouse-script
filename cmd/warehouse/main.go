package main

import (
	"flag"

	"github.com/warehouse/ai-service/internal/app"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "", "config file dir")
	flag.Parse()

	application := app.NewApplication(path)
	application.Run()
}
