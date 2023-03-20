package main

import (
	"log"
	"os"

	"fp/add"
	"fp/delete"
	"fp/query"
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
				Name:  "add",
				Usage: "Add a photo to the site",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "image",
						Aliases:  []string{"i"},
						Usage:    "Add `IMAGE_NAME` to site",
						Required: true,
					},

					&cli.StringSliceFlag{
						Name:     "tags",
						Aliases:  []string{"t"},
						Usage:    "Tags image with `TAGS` for later filtering",
						Required: true,
					},

					&cli.StringFlag{
						Name:     "publish_path",
						Aliases:  []string{"p"},
						EnvVars:  []string{"FP_PUBLISH_PATH"},
						Usage:    "Look for image inside directory `PUBLISH_PATH`",
						Required: true,
					},
				},

				Action: func(cCtx *cli.Context) error {
					return add.Action(cCtx)
				},
			},

			{
				Name:  "delete",
				Usage: "Delete a photo from the site",

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

			{
				Name:  "query",
				Usage: "Query for a specific image in the db",

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "image",
						Aliases:  []string{"i"},
						Usage:    "Query for `IMAGE_NAME` from db",
						Required: true,
					},
				},

				Action: func(cCtx *cli.Context) error {
					return query.Action(cCtx)
				},
			},

			{
				Name:  "validate",
				Usage: "Check that the entries in the db are well structured and pointing to valid blob storage locations",
				Action: func(cCtx *cli.Context) error {
					return validate.Action(cCtx)
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
