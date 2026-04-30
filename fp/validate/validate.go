package validate

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"

	"fp/imagedb"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"firebase.google.com/go/db"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:   "validate",
	Usage:  "Validate that all firebase database entries are well structured, point to proper blob entries, all images in database, bucket objects, and local image files are synced",
	Action: Action,
}

func Action(cCtx *cli.Context) error {
	log.Printf("VALIDATING")
	err := validate(imagedb.InitCloudClients())
	if err != nil {
		return err
	}

	log.Printf("VALID")
	return nil
}

func validate(dbClient *db.Client, s3Client *s3.Client) error {
	log.Printf("VALIDATING LOCAL IMAGE FILES")
	localImages, err := validateLocalImageDirectories()
	if err != nil {
		log.Printf("LOCAL IMAGE FILES INVALID")
		return err
	}
	log.Printf("LOCAL IMAGE FILES VALID")

	log.Printf("VALIDATING DATABASE")
	err = validateDatabase(dbClient, s3Client, localImages)
	if err != nil {
		log.Printf("DATABASE INVALID")
		return err
	}
	log.Printf("DATABASE VALID")

	return nil
}

func validateDatabase(dbClient *db.Client, s3Client *s3.Client, localImages []string) error {
	imagesRef := dbClient.NewRef("images")
	dbImages := map[string]imagedb.ImageEntry{}
	err := imagesRef.Get(context.Background(), &dbImages)
	if err != nil {
		return err
	}

	imageCountRef := dbClient.NewRef("imageCount")
	imageCount := 0
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}

	if imageCount != len(dbImages) {
		return fmt.Errorf("mismatch between /imageCount (%v) and number of database image entries (%v)", imageCount, len(dbImages))
	}

	keys := make([]string, 0, len(dbImages))
	for k := range dbImages {
		keys = append(keys, k)
	}
	diffImages := difference(keys, localImages)
	if len(diffImages) != 0 {
		return fmt.Errorf("The following images are in the database but not in the local directory: %v", diffImages)
	}
	diffImages = difference(localImages, keys)
	if len(diffImages) != 0 {
		return fmt.Errorf("The following images are in the local directory but not in the database: %v", diffImages)
	}

	var bucketName string
	dbClient.NewRef("bucket_name").Get(context.Background(), &bucketName)

	for name, entry := range dbImages {
		validateDatabaseImage(name, entry, bucketName, s3Client)
	}

	err = validateNoHangingBlobs(localImages, bucketName, s3Client)
	if err != nil {
		return err
	}

	return nil
}

func validateNoHangingBlobs(localImages []string, bucketName string, s3Client *s3.Client) error {
	output, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String("mini/"),
	})
	if err != nil {
		return err
	}

	var keys []string
	for _, object := range output.Contents {
		keys = append(keys, (*object.Key)[5:len(*object.Key)-4])
	}

	diffImages := difference(keys, localImages)
	if len(diffImages) != 0 {
		return fmt.Errorf("The following images are in s3 bucket but not in local directory: %v", diffImages)
	}

	return nil
}

func validateDatabaseImage(imageName string, dbEntry imagedb.ImageEntry, bucketName string, s3Client *s3.Client) error {
	if err := ensureUniqueTagForSet(dbEntry.Tags, imagedb.OrientationTags, "orientation"); err != nil {
		return fmt.Errorf("%v = INVALID: tags (%v) has %w", imageName, dbEntry.Tags, err)
	}

	if err := ensureUniqueTagForSet(dbEntry.Tags, imagedb.SpectrumTags, "spectrum"); err != nil {
		return fmt.Errorf("%v = INVALID: tags (%v) has %w", imageName, dbEntry.Tags, err)
	}

	if dbEntry.Priority < 0 {
		return fmt.Errorf("%v = INVALID: priority (%v) < 0", imageName, dbEntry.Priority)
	}

	fileInfo, err := os.Stat(os.Getenv("FP_PUBLISH_PATH") + "/mini/" + imageName + ".jpg")
	if err != nil {
		return err
	}
	if dbEntry.Size != fileInfo.Size() {
		return fmt.Errorf("%v = INVALID: database imageSize (%v) != file size (%v)", imageName, dbEntry.Size, fileInfo.Size())
	}

	if err := ensureValidBlobStorage(s3Client, bucketName, imageName, dbEntry); err != nil {
		return fmt.Errorf("%v = INVALID: invalid blob storage: %w", imageName, err)
	}

	return nil
}

func ensureValidBlobStorage(s3Client *s3.Client, bucketName string, imageName string, dbEntry imagedb.ImageEntry) error {
	attributes, err := s3Client.GetObjectAttributes(context.Background(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("mini/" + imageName + ".jpg"),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get object attributes from S3 bucket: %w", err)
	}

	if dbEntry.Size != *attributes.ObjectSize {
		return fmt.Errorf("difference between object size attribute (%d) and db imageSize (%d)", *attributes.ObjectSize, dbEntry.Size)
	}

	attributes, err = s3Client.GetObjectAttributes(context.Background(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("small/" + imageName + ".jpg"),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get object attributes from S3 bucket: %w", err)
	}

	attributes, err = s3Client.GetObjectAttributes(context.Background(), &s3.GetObjectAttributesInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("large/" + imageName + ".jpg"),
		ObjectAttributes: []types.ObjectAttributes{
			types.ObjectAttributesObjectSize,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get object attributes from S3 bucket: %w", err)
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

func validateLocalImageDirectories() ([]string, error) {
	// validate local image files all sync up with each other
	imageRootPath := os.Getenv("FP_PUBLISH_PATH")

	var largeImageNames []string
	dirEntries, err := os.ReadDir(imageRootPath + "/large")
	if err != nil {
		return nil, err
	}
	for _, image := range dirEntries {
		if image.IsDir() {
			continue
		}

		largeImageNames = append(largeImageNames, image.Name())
	}

	var smallImageNames []string
	dirEntries, err = os.ReadDir(imageRootPath + "/small")
	if err != nil {
		return nil, err
	}
	for _, image := range dirEntries {
		if image.IsDir() {
			continue
		}

		smallImageNames = append(smallImageNames, image.Name())
	}

	diffImages := difference(largeImageNames, smallImageNames)
	if len(diffImages) != 0 {
		return nil, fmt.Errorf("The following images are in large directory but not in small directory: %v", diffImages)
	}
	diffImages = difference(smallImageNames, largeImageNames)
	if len(diffImages) != 0 {
		return nil, fmt.Errorf("The following images are in small directory but not in large directory: %v", diffImages)
	}

	var miniImageNames []string
	dirEntries, err = os.ReadDir(imageRootPath + "/mini")
	if err != nil {
		return nil, err
	}
	for _, image := range dirEntries {
		if image.IsDir() {
			continue
		}

		miniImageNames = append(miniImageNames, image.Name())
	}

	diffImages = difference(largeImageNames, miniImageNames)
	if len(diffImages) != 0 {
		return nil, fmt.Errorf("The following images are in large directory but not in mini directory: %v", diffImages)
	}
	diffImages = difference(miniImageNames, largeImageNames)
	if len(diffImages) != 0 {
		return nil, fmt.Errorf("The following images are in mini directory but not in large directory: %v", diffImages)
	}

	for i, name := range largeImageNames {
		largeImageNames[i] = name[:len(name)-4]
	}

	return largeImageNames, nil
}

// difference returns the elements in `a` that are not in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
