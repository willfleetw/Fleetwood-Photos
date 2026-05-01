package clean

import (
	"fp/imagedb"
	"log"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "clean",
	Usage: "Clean the database by compacting the image priorities to remove holes created by deleting images. Preserves image order",

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	dbClient, _ := imagedb.InitCloudClients()

	log.Printf("CLEANING DATABASE")

	err := imagedb.CompactPriorities(dbClient)
	if err != nil {
		return err
	}

	log.Printf("DATABASE CLEANED")

	return nil
}
