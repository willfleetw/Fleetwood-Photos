package main

import (
	"log"
	"os"

	"fp/add"
	"fp/delete"
	"fp/query"
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
		delete.Command,
		store.Command,
		query.Command,
		validate.Command,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
