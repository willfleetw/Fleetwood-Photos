package delete

import (
	"context"
	"fp/imagedb"
	"log"

	"firebase.google.com/go/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	dbClient, s3Client := imagedb.InitCloudClients()

	log.Printf("DELETING: %v", imageName)
	err := delete(dbClient, s3Client, imageName)
	if err != nil {
		return err
	}
	log.Printf("DELETED: %v", imageName)
	return nil
}

func delete(dbClient *db.Client, s3Client *s3.Client, imageName string) error {
	imageRef := dbClient.NewRef("images/" + imageName)
	err := imageRef.Delete(context.Background())
	if err != nil {
		return err
	}

	imageCountRef := dbClient.NewRef("imageCount")
	var imageCount int64
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}
	err = imageCountRef.Set(context.Background(), imageCount-1)

	// TODO update image priorities

	var bucketName string
	dbClient.NewRef("bucket_name").Get(context.Background(), &bucketName)

	sizes := []string{"large", "small", "mini"}
	for _, size := range sizes {
		_, err := s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(size + "/" + imageName + ".jpg"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
