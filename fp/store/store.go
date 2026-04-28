package store

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "store",
	Usage: "Copy photos from a source directory to a new directory based on EXIF DateTime",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "source_dir",
			Aliases:  []string{"s"},
			EnvVars:  []string{"FP_STORE_SOURCE_DIR"},
			Usage:    "Copy images from `SOURCE_DIR`",
			Required: true,
		},

		&cli.StringFlag{
			Name:     "parent_destination_dir",
			Aliases:  []string{"p"},
			EnvVars:  []string{"FP_STORE_PARENT_DESTINATION_DIR"},
			Usage:    "Copy images to `PARENT_DESTINATION_DIR`/<DATE-TAKEN>",
			Required: true,
		},

		&cli.BoolFlag{
			Name:     "delete_originals",
			Aliases:  []string{"d"},
			Usage:    "Delete the images after copying them",
			Required: false,
			Value:    false,
		},
	},

	Action: Action,
}

var imageExtensions = []string{".RAF", ".JPEG", ".JPG", ".RAW", ".PNG"}

func Action(cCtx *cli.Context) error {
	sourceDir := cCtx.String("source_dir")
	parentDestDir := cCtx.String("parent_destination_dir")
	deleteOriginals := cCtx.Bool("delete_originals")

	dirEntries, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}

	images := []string{}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}

		fileExtension := strings.ToUpper(path.Ext(dirEntry.Name()))
		if slices.Contains(imageExtensions, fileExtension) {
			images = append(images, dirEntry.Name())
		}
	}

	var finalErr error
	for _, image := range images {
		errors.Join(finalErr, store_image(image, sourceDir, parentDestDir, deleteOriginals))
	}

	return finalErr
}

func store_image(image string, sourceDir string, parentDestDir string, deleteOriginals bool) error {
	imagePath := path.Join(sourceDir, image)

	srcInfo, err := os.Stat(imagePath)
	if err != nil {
		return fmt.Errorf("failed to get permission info for %s: %w", image, err)
	}

	source, err := os.Open(imagePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", image, err)
	}
	defer source.Close()

	exif, err := exif.Decode(source)
	if err != nil {
		return fmt.Errorf("failed to get datetime from EXIF for %s: %w", image, err)
	}

	dateTime, err := exif.DateTime()
	if err != nil {
		return fmt.Errorf("failed to get datetime from EXIF for %s: %w", image, err)
	}

	// Grabbing EXIF data reads from file, so reset offset back to start of file before copying
	_, err = source.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to set file offset to 0 for %s: %w", image, err)
	}

	dateDirPath := fmt.Sprintf("%d/%d_%.2d_%.2d", dateTime.Year(), dateTime.Year(), int(dateTime.Month()), dateTime.Day())

	dateTimeDir := path.Join(parentDestDir, dateDirPath)
	if err := os.MkdirAll(dateTimeDir, os.ModeDir); err != nil {
		return fmt.Errorf("failed to ensure %s exists before copying: %w", dateTimeDir, err)
	}

	destPath := path.Join(dateTimeDir, image)
	destination, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", destPath, err)
	}
	defer destination.Close()

	log.Printf("COPYING: %s -> %s", imagePath, destPath)
	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", image, destPath, err)
	}

	err = os.Chmod(destPath, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set permissions for %s: %w", destPath, err)
	}

	log.Printf("COPIED: %s -> %s", imagePath, destPath)

	if deleteOriginals {
		log.Printf("DELETING: %s", imagePath)
		err = os.Remove(imagePath)
		if err != nil {
			return fmt.Errorf("failed to delete original %s: %w", image, err)
		}
		log.Printf("DELETED: %s", imagePath)
	}

	return nil
}
