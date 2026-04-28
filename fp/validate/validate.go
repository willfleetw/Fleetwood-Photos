package validate

import (
	"context"
	"fmt"
	"log"

	"fp/imagedb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"cloud.google.com/go/storage"
	"firebase.google.com/go/db"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

var Command = &cli.Command{
	Name:   "validate",
	Usage:  "Check that the entries in the db are well structured and pointing to valid blob storage locations",
	Action: Action,
}

func Action(cCtx *cli.Context) error {
	return Validate(imagedb.InitFirebase())
}

func Validate(dbc *db.Client, bh *storage.BucketHandle) error {
	log.Print("VALIDATING DATABASE")
	err := validate(dbc, bh)
	if err != nil {
		log.Print("DATABASE: INVALID")
	} else {
		log.Print("DATABASE: VALID")
	}

	return err
}

func validate(dbc *db.Client, bh *storage.BucketHandle) error {
	imagesRef := dbc.NewRef("images")
	images := map[string]imagedb.ImageEntry{}
	err := imagesRef.Get(context.Background(), &images)
	if err != nil {
		return err
	}

	imageCountRef := dbc.NewRef("imageCount")
	imageCount := 0
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}

	if imageCount != len(images) {
		return fmt.Errorf("mismatch between /imageCount (%v) and number of image entries (%v)", imageCount, len(images))
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	client := s3.NewFromConfig(cfg)

	var bucketName string
	dbc.NewRef("bucket_name").Get(context.Background(), &bucketName)

	for name, entry := range images {
		if err = validateImage(client, bucketName, name, entry); err != nil {
			return err
		}
		log.Printf("%v = VALID", name)
	}

	return nil
}

func validateImage(client *s3.Client, bucketName string, name string, entry imagedb.ImageEntry) error {
	if err := ensureUniqueTagForSet(entry.Tags, imagedb.OrientationTags, "orientation"); err != nil {
		return fmt.Errorf("%v = INVALID: tags (%v) has %w", name, entry.Tags, err)
	}

	if err := ensureUniqueTagForSet(entry.Tags, imagedb.SpectrumTags, "spectrum"); err != nil {
		return fmt.Errorf("%v = INVALID: tags (%v) has %w", name, entry.Tags, err)
	}

	if entry.Priority < 0 {
		return fmt.Errorf("%v = INVALID: priority (%v) < 0", name, entry.Priority)
	}

	if entry.Size <= 0 {
		return fmt.Errorf("%v = INVALID: imageSize (%v) <= 0", name, entry.Size)
	}

	if err := ensureValidBlobStorage(client, bucketName, name, entry); err != nil {
		return fmt.Errorf("%v = INVALID: invalid blob storage: %w", name, err)
	}

	return nil
}

func ensureValidBlobStorage(client *s3.Client, bucketName string, name string, entry imagedb.ImageEntry) error {
	attributes, err := client.GetObjectAttributes(context.Background(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("mini/" + name + ".jpg"),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get object attributes from S3 bucket: %w", err)
	}

	if entry.Size != *attributes.ObjectSize {
		return fmt.Errorf("difference between object size attribute (%d) and db entry (%d)", *attributes.ObjectSize, entry.Size)
	}

	return nil
}

func ensureUniqueTagForSet(imageTags []string, tagSet []string, tagSetName string) error {
	foundOnceAlready := false
	for _, tag := range imageTags {
		if slices.Contains(tagSet, tag) {
			if foundOnceAlready {
				return fmt.Errorf("multiple %v tags", tagSetName)
			}
			foundOnceAlready = true
		}
	}

	if !foundOnceAlready {
		return fmt.Errorf("no %v tags", tagSetName)
	}

	return nil
}
