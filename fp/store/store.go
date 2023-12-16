package store

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
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

var fileTypeSuffixes = []string{".RAF", ".JPEG", ".JPG", ".RAW", ".PNG"}

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
		for _, type_suffix := range fileTypeSuffixes {
			if strings.HasSuffix(dirEntry.Name(), strings.ToUpper(type_suffix)) || strings.HasSuffix(dirEntry.Name(), strings.ToLower(type_suffix)) {
				images = append(images, dirEntry.Name())
			}
		}
	}

	imageErrors := make(chan error, len(images))
	for _, image := range images {
		go store_image(image, sourceDir, parentDestDir, deleteOriginals, imageErrors)
	}

	var finalErr error
	for range images {
		finalErr = errors.Join(finalErr, <-imageErrors)
	}

	return finalErr
}

func store_image(image string, sourceDir string, parentDestDir string, deleteOriginals bool, errors chan error) {
	imagePath := path.Join(sourceDir, image)
	dateTime, err := get_datetime(imagePath)
	if err != nil {
		errors <- fmt.Errorf("failed to get datetime from exif for %s: %w", image, err)
		return
	}

	dateTimeDir := path.Join(parentDestDir, dateTime)
	if err := os.MkdirAll(dateTimeDir, os.ModeDir); err != nil {
		errors <- fmt.Errorf("failed to ensure %s exists before copying: %w", dateTimeDir, err)
		return
	}

	destPath := path.Join(dateTimeDir, image)
	log.Printf("COPYING: %s -> %s", imagePath, destPath)
	input, err := os.ReadFile(imagePath)
	if err != nil {
		errors <- fmt.Errorf("failed to read %s for copying: %w", image, err)
		return
	}

	err = os.WriteFile(destPath, input, 0644)
	if err != nil {
		errors <- fmt.Errorf("failed to copy %s to %s: %w", image, destPath, err)
		return
	}
	log.Printf("COPIED: %s -> %s", imagePath, destPath)

	if deleteOriginals {
		log.Printf("DELETING: %s", imagePath)
		err = os.Remove(imagePath)
		if err != nil {
			errors <- fmt.Errorf("failed to delete original %s: %w", image, err)
			return
		}
		log.Printf("DELETED: %s", imagePath)
	}

	errors <- nil
}

func get_datetime(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	exif, err := exif.Decode(f)
	if err != nil {
		return "", err
	}

	dateTime, err := exif.DateTime()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d_%.2d_%.2d", dateTime.Year(), int(dateTime.Month()), dateTime.Day()), nil
}
