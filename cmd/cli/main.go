package main

import (
	"github.com/Utility-Gods/gottem/internal/app"
	"github.com/Utility-Gods/gottem/internal/cli"
)

func main() {
	myApp := app.NewApp()
	cli.RunCLI(myApp)
}
