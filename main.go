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
	"github.com/labstack/gommon/log"
	"github.com/martinlindhe/notify"
)

var (
	s3Bucket       = ""
	awsRegion      = ""
	awsConfigFile  = ""
	watchDirectory = ""
	s3Uploader     = &s3manager.Uploader{}
)

func initConfig() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	err := godotenv.Load(dir + "/.screenshot-uploader")
	if err != nil {
		log.Fatal("Error loading ~/.screenshot-uploader file")
	}

	s3Bucket = os.Getenv("S3_BUCKET")
	awsRegion = os.Getenv("AWS_REGION")
	awsConfigFile = os.Getenv("AWS_CONFIG_FILE")
	watchDirectory = os.Getenv("SCREENSHOT_LOCATION")

	creds := credentials.NewSharedCredentials(awsConfigFile, "default")

	conf := aws.Config{Region: aws.String(awsRegion), Credentials: creds}
	sess := session.New(&conf)
	s3Uploader = s3manager.NewUploader(sess)
}

func watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	if err := watcher.Add(watchDirectory); err != nil {
		log.Info("ERROR", err)
	}

	for {
		select {
		// watch for events
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				if path.Base(event.Name)[:1] == "." {
					continue
				}

				if filepath.Ext(event.Name) != ".png" {
					continue
				}

				log.Info("created file:", event.Name)

				s3Key := RandomString(5) + ".png"
				log.Info("s3 key: " + s3Key)

				file, _ := os.Open(event.Name)
				defer file.Close()

				log.Info("Uploading file to S3...")
				result, err := s3Uploader.Upload(&s3manager.UploadInput{
					Bucket:      aws.String(s3Bucket),
					Key:         aws.String(s3Key),
					Body:        file,
					ACL:         aws.String("public-read"),
					ContentType: aws.String("image/png"),
				})

				if err != nil {
					log.Info("s3 upload error", err)
					os.Exit(1)
				}

				log.Infof("Successfully uploaded %s to %s", event.Name, result.Location)

				clipboard.WriteAll(result.Location)
				notify.Notify("Screenshot Uploader", "", "Linked copied to clipboard", "")
			}

			// watch for errors
		case err := <-watcher.Errors:
			log.Info("ERROR", err)
		}
	}
}

func main() {
	initConfig()
	watchForChanges()
}
