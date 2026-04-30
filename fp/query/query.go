package query

import (
	"context"
	"log"

	"fp/imagedb"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "query",
	Usage: "Query firebase image database and S3 buckets",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "image",
			Aliases:  []string{"i"},
			Usage:    "Query database and bucket for `IMAGE_NAME`",
			Required: true,
		},
	},

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	dbc, _ := imagedb.InitCloudClients()

	imageName := cCtx.String("image")

	log.Printf("DATABASE: Querying for '%s'", imageName)

	imagesRef := dbc.NewRef("images")
	ref := imagesRef.Child(imageName)
	var imageEntry imagedb.ImageEntry
	ref.Get(context.Background(), &imageEntry)

	var bucketName string
	dbc.NewRef("bucket_name").Get(context.Background(), &bucketName)

	log.Printf("Priority: %d", imageEntry.Priority)
	log.Printf("Size: %d", imageEntry.Size)
	log.Printf("Tags: %s", imageEntry.Tags)
	log.Printf("Bucket Location: %s", bucketName+"/mini/"+imageName+".jpg")
	log.Printf("Bucket URL: %s", "https://"+bucketName+".s3.us-west-1.amazonaws.com/mini/"+imageName+".jpg")

	return nil
}
