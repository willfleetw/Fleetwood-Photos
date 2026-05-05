package add

import (
	"context"
	"fmt"
	"fp/imagedb"
	"image/jpeg"
	"log"
	"os"
	"strings"

	"firebase.google.com/go/db"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
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

	Action: Action,
}

func Action(cCtx *cli.Context) error {
	imageName := cCtx.String("image")
	publishPath := cCtx.String("publish_path")
	tags := cCtx.StringSlice("tags")
	for i, tag := range tags {
		tags[i] = strings.ToLower(tag)
	}

	dbClient, s3Client := imagedb.InitCloudClients()

	imagesRef := dbClient.NewRef("images")
	imageNames := map[string]bool{}
	err := imagesRef.GetShallow(context.Background(), &imageNames)
	if err != nil {
		return err
	}
	_, ok := imageNames[imageName]
	if ok {
		return fmt.Errorf("%v ALREADY EXISTS", imageName)
	}

	imageCountRef := dbClient.NewRef("imageCount")
	imageCount := 0
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}

	log.Printf("UPLOADING: %v", imageName)
	err = Upload(dbClient, s3Client, imageName, publishPath, tags, imageCount)
	if err != nil {
		log.Printf("NOT UPLOADED: %v", imageName)
	} else {
		imageCountRef.Set(context.Background(), imageCount+1)
		log.Printf("UPLOADED: %v", imageName)
	}

	return err
}

func Upload(
	dbClient *db.Client,
	s3Client *s3.Client,
	imageName string,
	publishPath string,
	tags []string,
	priority int,
) error {
	largeFilePath := fmt.Sprintf("%v/large/%v.jpg", publishPath, imageName)
	largeImageFile, err := os.Open(largeFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open LARGE image %v: %w", largeFilePath, err)
	}
	defer largeImageFile.Close()

	smallFilePath := fmt.Sprintf("%v/small/%v.jpg", publishPath, imageName)
	smallImageFile, err := os.Open(smallFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open SMALL image %v: %w", smallFilePath, err)
	}
	defer smallImageFile.Close()

	miniFilePath := fmt.Sprintf("%v/mini/%v.jpg", publishPath, imageName)
	miniFileStat, err := os.Stat(miniFilePath)
	if err != nil {
		return fmt.Errorf("error getting MINI image %v stat: %w", miniFilePath, err)
	}
	miniImageFile, err := os.Open(miniFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open MINI image %v: %w", miniFilePath, err)
	}
	defer miniImageFile.Close()

	var bucketName string
	dbClient.NewRef("bucket_name").Get(context.Background(), &bucketName)

	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String("large/" + imageName + ".jpg"),
		Body:        largeImageFile,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("error uploading LARGE image %s: %w", largeFilePath, err)
	}

	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String("small/" + imageName + ".jpg"),
		Body:        smallImageFile,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("error uploading SMALL image %s: %w", largeFilePath, err)
	}

	im, err := jpeg.DecodeConfig(miniImageFile)
	if err != nil {
		return fmt.Errorf("couldn't decode MINI image %v: %w", miniFilePath, err)
	}
	width, height := im.Width, im.Height

	ret, err := miniImageFile.Seek(0, 0)
	if err != nil || ret != 0 {
		return fmt.Errorf("seek error for MINI image %v: %w", miniFilePath, err)
	}

	_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String("mini/" + imageName + ".jpg"),
		Body:        miniImageFile,
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return fmt.Errorf("error uploading SMALL image %s: %w", largeFilePath, err)
	}

	orientation := "wide"
	if height > width {
		orientation = "tall"
	} else if height == width {
		orientation = "square"
	}

	tags = append(tags, orientation)
	fileRef := dbClient.NewRef(fmt.Sprintf("images/%v", imageName))
	imageEntry := imagedb.ImageEntry{
		Size:     miniFileStat.Size(),
		Priority: priority,
		Tags:     tags,
	}
	err = fileRef.Set(context.Background(), imageEntry)
	if err != nil {
		return fmt.Errorf("database set error for %v: %w", imageName, err)
	}

	return nil
}
