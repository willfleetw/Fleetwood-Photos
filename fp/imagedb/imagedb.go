package imagedb

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
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

func InitFirebase() (*db.Client, *storage.BucketHandle) {
	if _, envSet := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !envSet {
		log.Fatal("error: must be set GOOGLE_APPLICATION_CREDENTIALS before running")
	}

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

	return dbClient, bucketHandle
}
