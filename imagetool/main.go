package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/buckket/go-blurhash"
)

func initFirebase() (*db.Client, *storage.BucketHandle) {
	if _, envSet := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !envSet {
		log.Fatal("You must set GOOGLE_APPLICATION_CREDENTIALS before running\n")
	}

	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	dbClient, err := app.DatabaseWithURL(context.Background(), "https://fleetwood-photos-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("error getting database clien	t: %v\n", err)
	}
	storageClient, err := app.Storage(context.Background())
	if err != nil {
		log.Fatalf("error getting storage client: %v\n", err)
	}
	bucket, err := storageClient.Bucket("fleetwood-photos.appspot.com")
	if err != nil {
		log.Fatalf("error getting storage bucket: %v\n", err)
	}

	return dbClient, bucket
}

var largeFolderPath, smallFolderPath, miniFolderPath string

func uploadFile(fileInfo fs.FileInfo, dbClient *db.Client, bucket *storage.BucketHandle, wg *sync.WaitGroup) {
	defer wg.Done()

	// Check if fileInfo is a jpeg file, and that it exists in all three folders
	if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), ".jpg") {
		return
	}

	largeFilePath := filepath.Join(largeFolderPath, fileInfo.Name())
	if _, err := os.Stat(largeFilePath); os.IsNotExist(err) {
		fmt.Printf("large file %v does exist: %v\n", largeFilePath, err)
		return
	}
	smallFilePath := filepath.Join(smallFolderPath, fileInfo.Name())
	if _, err := os.Stat(smallFilePath); os.IsNotExist(err) {
		fmt.Printf("small file %v does exist: %v\n", smallFilePath, err)
		return
	}
	miniFilePath := filepath.Join(miniFolderPath, fileInfo.Name())

	// Check if file exists in storage, if so skip
	miniStorageObject := bucket.Object("images/mini/" + fileInfo.Name())
	_, err := miniStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up mini file %v in bucket: %v\n", fileInfo.Name(), err)
		return
	} else if err == nil {
		fmt.Printf("mini file %v already exists in bucket\n", fileInfo.Name())
		return
	}
	smallStorageObject := bucket.Object("images/small/" + fileInfo.Name())
	_, err = smallStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up small file %v in bucket: %v\n", fileInfo.Name(), err)
		return
	} else if err == nil {
		fmt.Printf("small file %v already exists in bucket\n", fileInfo.Name())
		return
	}
	largeStorageObject := bucket.Object("images/large/" + fileInfo.Name())
	_, err = largeStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up large file %v in bucket: %v\n", fileInfo.Name(), err)
		return
	} else if err == nil {
		fmt.Printf("largefile %v already exists in bucket\n", fileInfo.Name())
		return
	}

	smallImageFile, err := os.Open(smallFilePath)
	if err != nil {
		fmt.Printf("couldn't open image %v: %v\n", smallFilePath, err)
		return
	}
	defer smallImageFile.Close()
	smallFileStat, err := os.Stat(smallFilePath)
	if err != nil {
		fmt.Printf("error getting small image %v stat: %v\n", smallFilePath, err)
		return
	}

	largeImageFile, err := os.Open(largeFilePath)
	if err != nil {
		fmt.Printf("couldn't open image %v: %v\n", largeFilePath, err)
		return
	}
	defer largeImageFile.Close()
	largeFileStat, err := os.Stat(largeFilePath)
	if err != nil {
		fmt.Printf("error getting large image %v stat: %v\n", largeFilePath, err)
		return
	}

	//Decode image to get width, height, and calculate blur hash
	miniImageFile, err := os.Open(miniFilePath)
	if err != nil {
		fmt.Printf("couldn't open image %v: %v\n", miniFilePath, err)
		return
	}
	defer miniImageFile.Close()

	im, err := jpeg.DecodeConfig(miniImageFile)
	if err != nil {
		fmt.Printf("couldn't decode image %v: %v\n", miniFilePath, err)
		return
	}
	width, height := im.Width, im.Height

	ret, err := miniImageFile.Seek(0, 0)
	if err != nil {
		fmt.Printf("seek error: %v\n", err)
		return
	} else if ret != 0 {
		fmt.Printf("failed to seek to beginning of file %v\n", miniFilePath)
		return
	}

	loadedImage, err := jpeg.Decode(miniImageFile)
	if err != nil {
		fmt.Printf("decoding error for %v: %v\n", miniFilePath, err)
		return
	}
	blurHash, err := blurhash.Encode(4, 3, loadedImage) // 4x3 recomended by blurhash
	if err != nil {
		fmt.Printf("blurhash encoding error for %v: %v\n", miniFilePath, err)
		return
	}

	ret, err = miniImageFile.Seek(0, 0)
	if err != nil {
		fmt.Printf("seek error: %v\n", err)
		return
	} else if ret != 0 {
		fmt.Printf("failed to seek to beginning of file %v\n", miniFilePath)
		return
	}

	// Upload all three files to storage, and update database
	miniFileWriter := miniStorageObject.NewWriter(context.Background())
	miniFileWriter.ContentType = "image/jpeg"
	// recommened to set chunksize to slightly larger than filesize to avoid memory bloat
	miniFileWriter.ChunkSize = int(fileInfo.Size()) + 1024
	_, err = io.Copy(miniFileWriter, miniImageFile)
	if err != nil {
		fmt.Printf("failed copying file %v to storage writer: %v\n", miniFilePath, err)
		return
	}
	err = miniFileWriter.Close()
	if err != nil {
		fmt.Printf("error when writing file %v to storage: %v\n", miniFilePath, err)
		return
	}

	smallFileWriter := smallStorageObject.NewWriter(context.Background())
	smallFileWriter.ContentType = "image/jpeg"
	if smallFileStat.Size() < (16*1024*1024)-1024 { // default of 16MiB if file is too large
		smallFileWriter.ChunkSize = int(smallFileStat.Size()) + 1024
	}
	_, err = io.Copy(smallFileWriter, smallImageFile)
	if err != nil {
		fmt.Printf("failed copying file %v to storage writer: %v\n", smallFilePath, err)
		return
	}
	err = smallFileWriter.Close()
	if err != nil {
		fmt.Printf("error when writing file %v to storage: %v\n", smallFilePath, err)
		return
	}

	largeFileWriter := largeStorageObject.NewWriter(context.Background())
	largeFileWriter.ContentType = "image/jpeg"
	if largeFileStat.Size() < (16*1024*1024)-1024 { // default of 16MiB if file is too large
		largeFileWriter.ChunkSize = int(largeFileStat.Size()) + 1024
	}
	_, err = io.Copy(largeFileWriter, largeImageFile)
	if err != nil {
		fmt.Printf("failed copying file %v to storage writer: %v\n", largeFilePath, err)
		return
	}
	err = largeFileWriter.Close()
	if err != nil {
		fmt.Printf("error when writing file %v to storage: %v\n", largeFilePath, err)
		return
	}

	imageTitle := fileInfo.Name()[:len(fileInfo.Name())-4]
	fileDBRef := dbClient.NewRef("images/" + imageTitle)
	imageMetaData := make(map[string]interface{}, 0)
	imageMetaData["blurHash"] = blurHash
	imageMetaData["imageSize"] = fileInfo.Size()
	imageMetaData["width"] = width
	imageMetaData["height"] = height
	err = fileDBRef.Set(context.Background(), imageMetaData)
	if err != nil {
		fmt.Printf("database set error for %v: %v\n", imageTitle, err)
		return
	}
}

func main() {
	dbClient, bucket := initFirebase()
	// given a folder, containing "large/small/mini" subdirs, for each file that is in all three
	// 1. Check if it is already loaded into storage/DB. If so, skip
	// 2. Calculate needed information from mini file, seek back to front of file
	// 3. Load all three files to storage and calculated info to DB

	folderPath := filepath.Clean(os.Args[1])
	largeFolderPath = filepath.Join(folderPath, "large")
	if _, err := os.Stat(largeFolderPath); os.IsNotExist(err) {
		log.Fatalf("directory %v does exist: %v", largeFolderPath, err)
	}
	smallFolderPath = filepath.Join(folderPath, "small")
	if _, err := os.Stat(smallFolderPath); os.IsNotExist(err) {
		log.Fatalf("directory %v does exist: %v", smallFolderPath, err)
	}
	miniFolderPath = filepath.Join(folderPath, "mini")
	if _, err := os.Stat(miniFolderPath); os.IsNotExist(err) {
		log.Fatalf("directory %v does exist: %v", miniFolderPath, err)
	}

	// we need to list over each file in mini, and check for existence in other folders
	files, err := ioutil.ReadDir(miniFolderPath)
	if err != nil {
		log.Fatalf("failed to read %v: %v\n", miniFolderPath, err)
	}

	var wg sync.WaitGroup
	for _, fi := range files {
		wg.Add(1)
		go uploadFile(fi, dbClient, bucket, &wg)
	}
	wg.Wait()

	// Now we need to update total image count
	imagesRef := dbClient.NewRef("images")
	imageNames := make(map[string]interface{}, 0)
	err = imagesRef.GetShallow(context.Background(), &imageNames)
	if err != nil {
		log.Fatalf("failed to get total image count: %v\n", err)
	}
	imageCountRef := dbClient.NewRef("imageCount")
	err = imageCountRef.Set(context.Background(), len(imageNames))
	if err != nil {
		log.Fatalf("failed to set new image count %v: %v", len(imageNames), err)
	}
}
