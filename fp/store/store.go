package store

import (
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
	Usage: "Copy photos from `SOURCE_DIR`, and place them in `DESTINATION_DIR`/<DATE-TAKEN>",

	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "source_dir",
			Aliases:  []string{"s"},
			Usage:    "Copy images from `SOURCE_DIR`",
			Required: true,
		},

		&cli.StringFlag{
			Name:     "destination_dir",
			Aliases:  []string{"d"},
			EnvVars:  []string{"FP_STORE_DESTINATION_DIR"},
			Usage:    "Copy images to `DESTINATION_DIR`/<DATE-TAKEN>",
			Required: true,
		},

		&cli.BoolFlag{
			Name:     "delete_originals",
			Aliases:  []string{"c"},
			Usage:    "Delete the images after copying them",
			Required: false,
			Value:    false,
		},
	},

	Action: Action,
}

var file_type_suffixes = []string{".RAF", ".JPEG", ".JPG", ".RAW", ".PNG"}

func Action(cCtx *cli.Context) error {
	source_dir := cCtx.String("source_dir")
	destination_dir := cCtx.String("destination_dir")
	delete_originals := cCtx.Bool("delete_originals")

	dirEntries, err := os.ReadDir(source_dir)
	if err != nil {
		return err
	}

	images := []string{}
	for _, dirEntry := range dirEntries {
		if dirEntry.IsDir() {
			continue
		}
		for _, type_suffix := range file_type_suffixes {
			if strings.HasSuffix(dirEntry.Name(), strings.ToUpper(type_suffix)) || strings.HasSuffix(dirEntry.Name(), strings.ToLower(type_suffix)) {
				images = append(images, dirEntry.Name())
			}
		}
	}

	for _, image := range images {
		imagePath := path.Join(source_dir, image)
		dateTime, err := get_datetime(imagePath)
		if err != nil {
			return fmt.Errorf("failed to get datetime from exif for %s: %w", image, err)
		}

		dateTimeDirPath := path.Join(destination_dir, dateTime)
		if err := os.MkdirAll(dateTimeDirPath, os.ModeDir); err != nil {
			return fmt.Errorf("failed to ensure %s exists before copying: %w", dateTimeDirPath, err)
		}

		destination_imagePath := path.Join(dateTimeDirPath, image)
		log.Printf("COPYING: %s -> %s", imagePath, destination_imagePath)
		input, err := os.ReadFile(imagePath)
		if err != nil {
			return fmt.Errorf("failed to read %s for copying: %w", image, err)
		}

		err = os.WriteFile(destination_imagePath, input, 0644)
		if err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", image, destination_imagePath, err)
		}
		log.Printf("COPIED: %s -> %s", imagePath, destination_imagePath)

		if delete_originals {
			log.Printf("DELETING: %s", imagePath)
			err = os.Remove(imagePath)
			if err != nil {
				return fmt.Errorf("failed to delete original %s: %w", image, err)
			}
			log.Printf("DELETED: %s", imagePath)
		}
	}

	return nil
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
