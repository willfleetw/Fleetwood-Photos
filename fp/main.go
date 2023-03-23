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

func main() {
	app := &cli.App{
		Name:                 "fp",
		Usage:                "manage images for https://fleetwood.photos",
		EnableBashCompletion: true,

		Commands: []*cli.Command{
			add.Command,
			delete.Command,
			store.Command,
			query.Command,
			validate.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
