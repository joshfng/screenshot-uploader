package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/martinlindhe/notify"
)

var (
	s3Bucket       = ""
	s3Host         = ""
	awsRegion      = ""
	awsProfile     = ""
	awsConfigFile  = ""
	watchDirectory = ""
	s3Uploader     = &s3manager.Uploader{}
)

func initConfig() {
	user, _ := user.Current()
	directory := user.HomeDir
	err := godotenv.Load(directory + "/.config/screenshot-uploader")
	if err != nil {
		panic("Error loading ~/.config/screenshot-uploader file")
	}

	s3Bucket = os.Getenv("S3_BUCKET")
	s3Host = os.Getenv("S3_HOST")
	awsRegion = os.Getenv("AWS_REGION")
	awsProfile = os.Getenv("AWS_PROFILE")
	awsConfigFile = os.Getenv("AWS_CONFIG_FILE")
	watchDirectory = os.Getenv("SCREENSHOT_LOCATION")

	fmt.Printf("Watching for changes in %s\n", watchDirectory)

	creds := credentials.NewSharedCredentials(awsConfigFile, awsProfile)

	config := aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: creds,
	}
	session := session.New(&config)
	s3Uploader = s3manager.NewUploader(session)
}

func uploadScreenshot(filePath string) {
	ext := filepath.Ext(filePath)

	if ext != ".png" && ext != ".jpg" && ext != ".mov" {
		return
	}

	mimeType := ""

	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".jpg":
		mimeType = "image/jpeg"
	case ".mov":
		mimeType = "video/quicktime"
	}

	s3Key := RandomString(5) + ext
	fmt.Println("s3 key: " + s3Key)

	file, _ := os.Open(filePath)

	fmt.Println("Uploading file to S3...")
	result, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ACL:         aws.String("public-read"),
		ContentType: aws.String(mimeType),
	})

	file.Close()

	if err != nil {
		fmt.Println("s3 upload error", err)
		return
	}

	fmt.Printf("Successfully uploaded %s to %s\n", filePath, result.Location)

	url := result.Location
	if s3Host != "" {
		url = s3Host + "/" + s3Key
	}

	sendNotification(url)
}

func sendNotification(url string) {
	clipboard.WriteAll(url)
	notify.Notify("Screenshot Uploaded!", "", "Linked copied to clipboard", "")
}

func watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer watcher.Close()

	if err := watcher.Add(watchDirectory); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for {
		select {
		case event := <-watcher.Events:
			// fmt.Printf("event: %v\n", event)
			switch event.Op {
			case fsnotify.Write, fsnotify.Create:
				if path.Base(event.Name)[:1] == "." {
					continue
				}

				if path.Base(event.Name)[:1] == ".." {
					continue
				}

				uploadScreenshot(event.Name)
			}

		case err := <-watcher.Errors:
			// if err == nil {
			//     panic("unexpected nil err")
			// }
			// return
			panic(err)
		}
	}
}

func main() {
	initConfig()
	watchForChanges()
}
