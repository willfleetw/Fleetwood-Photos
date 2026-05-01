package imagedb

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageEntry struct {
	Size     int64    `json:"imageSize"`
	Priority int      `json:"priority"`
	Tags     []string `json:"tags"`
}

var (
	OrientationTags = []string{"wide", "tall", "square"}
	SpectrumTags    = []string{"blackandwhite", "color"}
)

func InitCloudClients() (*db.Client, *s3.Client) {
	if _, envSet := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !envSet {
		log.Fatal("error: must be set GOOGLE_APPLICATION_CREDENTIALS before running")
	}

	fbApp, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing firebase: %v", err)
	}

	dbClient, err := fbApp.DatabaseWithURL(context.Background(), "https://fleetwood-photos-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("error getting database client: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	s3Client := s3.NewFromConfig(cfg)

	return dbClient, s3Client
}

func CompactPriorities(dbClient *db.Client) error {
	ref := dbClient.NewRef("images")
	query := ref.OrderByChild("priority")

	queryNodes, err := query.GetOrdered(context.Background())
	if err != nil {
		return err
	}

	for idx, entry := range queryNodes {
		var image ImageEntry
		err := entry.Unmarshal(&image)
		if err != nil {
			return err
		}

		if image.Priority != idx {
			log.Printf("UPDATING %v priority from %v -> %v", entry.Key(), image.Priority, idx)
		}
		image.Priority = idx

		imageRef := dbClient.NewRef(fmt.Sprintf("images/%s", entry.Key()))
		err = imageRef.Set(context.Background(), image)
		if err != nil {
			return err
		}
	}

	return nil
}
