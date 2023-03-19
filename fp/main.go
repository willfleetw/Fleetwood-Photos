package main

import (
	"log"
	"os"

	"fp/delete"
	"fp/validate"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "fp",
		Usage:                "manage images for https://fleetwood.photos",
		EnableBashCompletion: true,

		Commands: []*cli.Command{
			{
				Name:  "validate",
				Usage: "Check that the entries in the db are well structured and pointing to valid blob storage locations",
				Action: func(cCtx *cli.Context) error {
					return validate.Action(cCtx)
				},
			},

			{
				Name:  "delete",
				Usage: "Delete a photo from both the db and blob storage, thus removing from site",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "image",
						Aliases:  []string{"i"},
						Usage:    "Delete `IMAGE_NAME` from site",
						Required: true,
					},
				},

				Action: func(cCtx *cli.Context) error {
					return delete.Action(cCtx)
				},
			},
		},
	}

	if _, envSet := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !envSet {
		log.Fatal("You must set GOOGLE_APPLICATION_CREDENTIALS before running")
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
