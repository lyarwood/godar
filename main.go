package main

import (
	"github.com/lyarwood/godar/cmd"
)

// Version information - set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	cmd.Execute()
}
