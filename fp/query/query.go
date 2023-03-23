package query

import (
	"context"
	"fmt"
	"log"

	"fp/imagedb"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
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

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	imageName := cCtx.String("image")

	dbc, _ := imagedb.InitFirebase()
	imageRef := dbc.NewRef(fmt.Sprintf("images/%v", imageName))
	imageEntry := imagedb.ImageEntry{}
	err := imageRef.Get(context.Background(), &imageEntry)
	if err != nil {
		return err
	}

	log.Printf("%v: %+v", imageName, imageEntry)

	return nil
}
