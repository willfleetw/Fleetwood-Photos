package delete

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
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

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	imageName := cCtx.String("image")

	fbApp, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	dbClient, err := fbApp.DatabaseWithURL(context.Background(), "https://fleetwood-photos-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("error getting database client: %v", err)
	}

	storageClient, err := fbApp.Storage(context.Background())
	if err != nil {
		log.Fatalf("error getting storage client: %v", err)
	}
	bucketHandle, err := storageClient.Bucket("fleetwood-photos.appspot.com")
	if err != nil {
		log.Fatalf("error getting storage bucket handle: %v", err)
	}

	log.Printf("DELETING: %v", imageName)
	err = delete(dbClient, bucketHandle, imageName)
	if err != nil {
		log.Printf("NOT DELETED: %v", imageName)
	} else {
		log.Printf("DELETED: %v", imageName)
	}

	return err
}

func delete(dbc *db.Client, bh *storage.BucketHandle, imageName string) error {
	imageRef := dbc.NewRef(fmt.Sprintf("images/%v", imageName))
	err := imageRef.Delete(context.Background())
	if err != nil {
		return err
	}

	sizes := []string{"large", "small", "mini"}
	for _, size := range sizes {
		oh := bh.Object(fmt.Sprintf("images/%v/%v.jpg", size, imageName))
		err = oh.Delete(context.Background())
		if err != nil {
			return err
		}
	}

	imageCountRef := dbc.NewRef("imageCount")
	imageCount := 0
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}
	err = imageCountRef.Set(context.Background(), imageCount-1)
	return err
}
