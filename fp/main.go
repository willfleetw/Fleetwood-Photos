package main

import (
	"log"
	"os"

	"fp/add"
	"fp/clean"
	"fp/delete"
	"fp/store"
	"fp/validate"

	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:                 "fp",
	Usage:                "Manage images for https://fleetwood.photos",
	EnableBashCompletion: true,

	Commands: []*cli.Command{
		add.Command,
		clean.Command,
		delete.Command,
		store.Command,
		validate.Command,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
