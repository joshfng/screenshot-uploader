# Screenshot Uploader

A simple go program for MacOS that watches for screenshots and uploads them to s3. A random, short object key is generated for the image and the s3 url is automatically copied to your clipboard, along with a native notification when the upload completes

![Demo](demo.gif)

## Setup

This program assumes you have an aws config file. Next create a file in your home directory called .screenshot-uploader:

```.env
AWS_REGION=us-east-1
AWS_CONFIG_FILE=/full/path/to/.aws/config
S3_BUCKET=i.joshfrye.com

# and new .png added here will be uploaded
SCREENSHOT_LOCATION=/Users/josh/Desktop
```

## TODO

- Menu bar item
