package validate

import (
	"context"
	"fmt"
	"log"

	"fp/imagedb"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func Action(cCtx *cli.Context) error {
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

	log.Print("validating database")
	err = validate(dbClient, bucketHandle)
	if err != nil {
		log.Print("database: INVALID")
	} else {
		log.Print("database: VALID")
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

	for name, entry := range images {
		if err = validateImage(bh, name, entry); err != nil {
			return err
		}
		log.Printf("\t%v = VALID", name)
	}

	return nil
}

func validateImage(bh *storage.BucketHandle, name string, entry imagedb.ImageEntry) error {
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

	if err := ensureValidBlobStorage(bh, name, entry); err != nil {
		return fmt.Errorf("%v = INVALID: invalid blob storage: %w", name, err)
	}

	return nil
}

func ensureValidBlobStorage(bh *storage.BucketHandle, name string, entry imagedb.ImageEntry) error {
	objectHandle := bh.Object(fmt.Sprintf("images/mini/%v.jpg", name))
	miniImageAttr, err := objectHandle.Attrs(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get MINI file: %w", err)
	}

	if miniImageAttr.Size != entry.Size {
		return fmt.Errorf("difference between MINI.Size attribute (%v) and db entry (%v)", miniImageAttr.Size, entry.Size)
	}

	objectHandle = bh.Object(fmt.Sprintf("images/small/%v.jpg", name))
	_, err = objectHandle.Attrs(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get SMALL file: %w", err)
	}

	objectHandle = bh.Object(fmt.Sprintf("images/large/%v.jpg", name))
	_, err = objectHandle.Attrs(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get LARGE file: %w", err)
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
