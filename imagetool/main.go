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
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"github.com/buckket/go-blurhash"
	"github.com/manifoldco/promptui"
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

var mu sync.Mutex
var maxPriority int

// Check if file already exists in DB
func checkFileExistance(fileInfo fs.FileInfo, dbClient *db.Client, bucket *storage.BucketHandle, wg *sync.WaitGroup, ch chan string) {
	defer wg.Done()

	// Check if fileInfo is a jpeg file, and that it exists in all three folders
	if fileInfo.IsDir() || !strings.HasSuffix(fileInfo.Name(), ".jpg") {
		return
	}

	largeFilePath := filepath.Join(largeFolderPath, fileInfo.Name())
	if _, err := os.Stat(largeFilePath); os.IsNotExist(err) {
		fmt.Printf("large file %v does not exist: %v\n", largeFilePath, err)
		return
	}
	smallFilePath := filepath.Join(smallFolderPath, fileInfo.Name())
	if _, err := os.Stat(smallFilePath); os.IsNotExist(err) {
		fmt.Printf("small file %v does not exist: %v\n", smallFilePath, err)
		return
	}
	miniFilePath := filepath.Join(miniFolderPath, fileInfo.Name())
	if _, err := os.Stat(miniFilePath); os.IsNotExist(err) {
		fmt.Printf("mini file %v does not exist: %v\n", miniFilePath, err)
		return
	}

	// Check if image exists in DB
	imageTitle := fileInfo.Name()[:len(fileInfo.Name())-4] // remove .jpg from name
	ref := dbClient.NewRef("images/" + imageTitle)
	imageData := make(map[string]interface{}, 0)
	err := ref.Get(context.Background(), &imageData)
	if err != nil {
		fmt.Printf("failed to read db for %v: %v\n", imageTitle, err)
		return
	}

	// image does not exist in DB, so we upload
	if len(imageData) == 0 {
		ch <- imageTitle
	} else { // image does exist, so take oportunity to read priority
		mu.Lock()
		imagePriority := int(imageData["priority"].(float64))
		if imagePriority > maxPriority {
			maxPriority = imagePriority
		}
		mu.Unlock()
	}
}

// given a folder, containing "large/small/mini" subdirs, for each file that is in all three
// 1. Check if it is already loadd into DB. If so, skip
// 2. Calculate needed information from mini file, seek back to front of file
// 3. Load all three files to storage and calculated info to DB
func uploadFiles(folderPath string, dbClient *db.Client, bucket *storage.BucketHandle) error {
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

	files, err := ioutil.ReadDir(miniFolderPath)
	if err != nil {
		log.Fatalf("failed to list files in  %v: %v\n", miniFolderPath, err)
	}

	var wg sync.WaitGroup
	ch := make(chan string, len(files))
	for _, fi := range files {
		wg.Add(1)
		go checkFileExistance(fi, dbClient, bucket, &wg, ch)
	}
	wg.Wait()
	close(ch)
	maxPriority++

	// ch now only contains images that are not in DB
	// Now we must ask the user which tags to apply to each image
	filesToAdd := make(map[string][]string, 0)
	for fileName := range ch {
		switch fileName {
		case "":
			continue
		default:
			// Prompt user somehow, and get tags
			var tags []string
			fmt.Printf("Please input tags for %v:\n", fileName)
			for {
				prompt := promptui.Prompt{
					Label: fmt.Sprintf("Tag [%v]: ", len(tags)),
				}
				result, err := prompt.Run()
				if err != nil {
					return err
				}
				if result == "" {
					break
				} else {
					tags = append(tags, result)
				}
			}
			filesToAdd[fileName] = tags
		}
	}

	for fileName, tags := range filesToAdd {
		wg.Add(1)
		go uploadFile(fileName, tags, maxPriority, dbClient, bucket, &wg)
	}
	wg.Wait()

	return nil
}

// Given a fileName, upload large/small/mini files to storage, parse and upload info to DB
func uploadFile(fileName string, tags []string, priority int, dbClient *db.Client, bucket *storage.BucketHandle, wg *sync.WaitGroup) {
	defer wg.Done()

	// Check if file exists in storage, if so skip
	miniStorageObject := bucket.Object("images/mini/" + fileName + ".jpg")
	_, err := miniStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up mini file %v in bucket: %v\n", fileName, err)
		return
	} else if err == nil {
		fmt.Printf("mini file %v already exists in bucket\n", fileName)
		return
	}
	smallStorageObject := bucket.Object("images/small/" + fileName + ".jpg")
	_, err = smallStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up small file %v in bucket: %v\n", fileName, err)
		return
	} else if err == nil {
		fmt.Printf("small file %v already exists in bucket\n", fileName)
		return
	}
	largeStorageObject := bucket.Object("images/large/" + fileName + ".jpg")
	_, err = largeStorageObject.Attrs(context.Background())
	if err != nil && err != storage.ErrObjectNotExist {
		fmt.Printf("error looking up large file %v in bucket: %v\n", fileName, err)
		return
	} else if err == nil {
		fmt.Printf("largefile %v already exists in bucket\n", fileName)
		return
	}

	largeFilePath := filepath.Join(largeFolderPath, fileName+".jpg")
	if _, err := os.Stat(largeFilePath); os.IsNotExist(err) {
		fmt.Printf("large file %v does not exist: %v\n", largeFilePath, err)
		return
	}
	smallFilePath := filepath.Join(smallFolderPath, fileName+".jpg")
	if _, err := os.Stat(smallFilePath); os.IsNotExist(err) {
		fmt.Printf("small file %v does not exist: %v\n", smallFilePath, err)
		return
	}
	miniFilePath := filepath.Join(miniFolderPath, fileName+".jpg")
	if _, err := os.Stat(miniFilePath); os.IsNotExist(err) {
		fmt.Printf("mini file %v does not exist: %v\n", miniFilePath, err)
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
	miniFileStat, err := os.Stat(miniFilePath)
	if err != nil {
		fmt.Printf("error getting mini image %v stat: %v\n", miniFilePath, err)
		return
	}

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
	miniFileWriter.ChunkSize = int(miniFileStat.Size()) + 1024
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

	fileDBRef := dbClient.NewRef("images/" + fileName)
	imageMetaData := make(map[string]interface{}, 0)
	imageMetaData["blurHash"] = blurHash
	imageMetaData["imageSize"] = miniFileStat.Size()
	imageMetaData["width"] = width
	imageMetaData["height"] = height
	imageMetaData["priority"] = priority
	imageMetaData["tags"] = tags
	err = fileDBRef.Set(context.Background(), imageMetaData)
	if err != nil {
		fmt.Printf("database set error for %v: %v\n", fileName, err)
		return
	}

	fmt.Printf("uploaded %v\n", fileName)
}

func main() {
	dbClient, bucket := initFirebase()
	folderPath := filepath.Clean(os.Args[1])
	uploadFiles(folderPath, dbClient, bucket)

	// Now we need to update total image count
	imagesRef := dbClient.NewRef("images")
	imageNames := make(map[string]interface{}, 0)
	err := imagesRef.GetShallow(context.Background(), &imageNames)
	if err != nil {
		log.Fatalf("failed to get total image count: %v\n", err)
	}
	imageCountRef := dbClient.NewRef("imageCount")
	err = imageCountRef.Set(context.Background(), len(imageNames))
	if err != nil {
		log.Fatalf("failed to set new image count to %v: %v", len(imageNames), err)
	}
}
