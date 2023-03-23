package add

import (
	"context"
	"fmt"
	"fp/imagedb"
	"image/jpeg"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"firebase.google.com/go/db"
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

	dbc, bh := imagedb.InitFirebase()

	imagesRef := dbc.NewRef("images")
	imageNames := map[string]bool{}
	err := imagesRef.GetShallow(context.Background(), &imageNames)
	if err != nil {
		return err
	}
	_, ok := imageNames[imageName]
	if ok {
		return fmt.Errorf("%v ALREADY EXISTS", imageName)
	}

	imageCountRef := dbc.NewRef("imageCount")
	imageCount := 0
	err = imageCountRef.Get(context.Background(), &imageCount)
	if err != nil {
		return err
	}

	log.Printf("UPLOADING: %v", imageName)
	err = Upload(dbc, bh, imageName, publishPath, tags, imageCount)
	if err != nil {
		log.Printf("NOT UPLOADED: %v", imageName)
	} else {
		imageCountRef.Set(context.Background(), imageCount+1)
		log.Printf("UPLOADED: %v", imageName)
	}

	return err
}

func Upload(
	dbc *db.Client,
	bh *storage.BucketHandle,
	imageName string,
	publishPath string,
	tags []string,
	priority int,
) error {
	largeStorageObject := bh.Object(fmt.Sprintf("images/large/%v.jpg", imageName))
	largeFilePath := fmt.Sprintf("%v/large/%v.jpg", publishPath, imageName)
	if _, err := os.Stat(largeFilePath); os.IsNotExist(err) {
		return fmt.Errorf("LARGE file %v does not exist: %w", largeFilePath, err)
	}

	smallStorageObject := bh.Object(fmt.Sprintf("images/small/%v.jpg", imageName))
	smallFilePath := fmt.Sprintf("%v/small/%v.jpg", publishPath, imageName)
	if _, err := os.Stat(smallFilePath); os.IsNotExist(err) {
		return fmt.Errorf("SMALL file %v does not exist: %w", smallFilePath, err)
	}

	miniStorageObject := bh.Object(fmt.Sprintf("images/mini/%v.jpg", imageName))
	miniFilePath := fmt.Sprintf("%v/mini/%v.jpg", publishPath, imageName)
	if _, err := os.Stat(miniFilePath); os.IsNotExist(err) {
		return fmt.Errorf("MINI file %v does not exist: %w", miniFilePath, err)
	}

	largeImageFile, err := os.Open(largeFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open LARGE image %v: %w", largeFilePath, err)
	}
	defer largeImageFile.Close()

	smallImageFile, err := os.Open(smallFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open SMALL image %v: %w", smallFilePath, err)
	}
	defer smallImageFile.Close()

	miniImageFile, err := os.Open(miniFilePath)
	if err != nil {
		return fmt.Errorf("couldn't open MINI image %v: %w", miniFilePath, err)
	}
	defer miniImageFile.Close()

	largeFileStat, err := os.Stat(largeFilePath)
	if err != nil {
		return fmt.Errorf("error getting LARGE image %v stat: %w", largeFilePath, err)
	}

	smallFileStat, err := os.Stat(smallFilePath)
	if err != nil {
		return fmt.Errorf("error getting SMALL image %v stat: %w", smallFilePath, err)
	}

	miniFileStat, err := os.Stat(miniFilePath)
	if err != nil {
		return fmt.Errorf("error getting MINI image %v stat: %w", miniFilePath, err)
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

	largeFileWriter := largeStorageObject.NewWriter(context.Background())
	largeFileWriter.ContentType = "image/jpeg"
	if largeFileStat.Size() < (16*1024*1024)-1024 { // default of 16MiB if file is too large
		largeFileWriter.ChunkSize = int(largeFileStat.Size()) + 1024
	}
	_, err = io.Copy(largeFileWriter, largeImageFile)
	if err != nil {
		return fmt.Errorf("failed copying LARGE file %v to storage writer: %w", largeFilePath, err)
	}
	err = largeFileWriter.Close()
	if err != nil {
		return fmt.Errorf("failed when writing LARGE file %v to storage writer: %w", largeFilePath, err)
	}

	smallFileWriter := smallStorageObject.NewWriter(context.Background())
	smallFileWriter.ContentType = "image/jpeg"
	if smallFileStat.Size() < (16*1024*1024)-1024 { // default of 16MiB if file is too large
		smallFileWriter.ChunkSize = int(smallFileStat.Size()) + 1024
	}
	_, err = io.Copy(smallFileWriter, smallImageFile)
	if err != nil {
		return fmt.Errorf("failed copying SMALL file %v to storage writer: %w", smallFilePath, err)
	}
	err = smallFileWriter.Close()
	if err != nil {
		return fmt.Errorf("failed when writing SMALL file %v to storage writer: %w", smallFilePath, err)
	}

	miniFileWriter := miniStorageObject.NewWriter(context.Background())
	miniFileWriter.ContentType = "image/jpeg"
	if miniFileStat.Size() < (16*1024*1024)-1024 { // default of 16MiB if file is too large
		miniFileWriter.ChunkSize = int(miniFileStat.Size()) + 1024
	}
	_, err = io.Copy(miniFileWriter, miniImageFile)
	if err != nil {
		return fmt.Errorf("failed copying MINI file %v to storage writer: %w", miniFilePath, err)
	}
	err = miniFileWriter.Close()
	if err != nil {
		return fmt.Errorf("failed when writing MINI file %v to storage writer: %w", miniFilePath, err)
	}

	orientation := "wide"
	if height > width {
		orientation = "tall"
	} else if height == width {
		orientation = "square"
	}

	tags = append(tags, orientation)
	fileRef := dbc.NewRef(fmt.Sprintf("images/%v", imageName))
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
