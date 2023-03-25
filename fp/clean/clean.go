package clean

import (
	"context"
	"fmt"
	"log"

	"fp/imagedb"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "clean",
	Usage: "Clean the database",

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	dbc, _ := imagedb.InitFirebase()
	ref := dbc.NewRef("images")
	query := ref.OrderByChild("priority")

	qbnds, err := query.GetOrdered(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for idx, entry := range qbnds {
		image := imagedb.ImageEntry{}
		err := entry.Unmarshal(&image)
		if err != nil {
			log.Fatal(err)
		}
		image.Priority = idx
		imageRef := dbc.NewRef(fmt.Sprintf("images/%s", entry.Key()))
		err = imageRef.Set(context.Background(), image)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
