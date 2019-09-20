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
	s3Host         = ""
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
	s3Host = os.Getenv("S3_HOST")
	awsRegion = os.Getenv("AWS_REGION")
	awsConfigFile = os.Getenv("AWS_CONFIG_FILE")
	watchDirectory = os.Getenv("SCREENSHOT_LOCATION")

	creds := credentials.NewSharedCredentials(awsConfigFile, "default")

	conf := aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: creds,
	}
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

				url := result.Location
				if s3Host != "" {
					url = s3Host + "/" + s3Key
				}

				clipboard.WriteAll(url)
				notify.Notify("Screenshot Uploader", "", "Linked copied to clipboard", "")

				file.Close()
			}

		case err := <-watcher.Errors:
			log.Info("ERROR", err)
		}
	}
}

func main() {
	initConfig()
	watchForChanges()
}
