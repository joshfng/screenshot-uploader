# Screenshot Uploader

A simple go program for MacOS that watches for screenshots or screen recordings and uploads them to s3. A random, short object key is generated for the image and the s3 url is automatically copied to your clipboard, along with a native notification when the upload completes

![Demo](demo.gif)

## Setup

This program assumes you have an aws config file. Next create a file in your home directory called .screenshot-uploader:

```.env
AWS_REGION=us-east-1
AWS_PROFILE=default
AWS_CONFIG_FILE=/full/path/to/.aws/config
S3_BUCKET=i.mydomain.com

# leave blank to use default s3 url, only set if you have a cname record
S3_HOST=http://i.mydomain.com

# any new .png or .mov added here will be uploaded
SCREENSHOT_LOCATION=/Users/j/Desktop
```

Download a release and place it somewhere in your path, then run `screenshot-uploader`. Any new pngs or movs created in your watch directory will be uploaded.

If you want the program to always run, download `com.screenshot.uploader.plist`, edit the path to the binary, and run `launchctl load com.screenshot.uploader.plist`.

## TODO

- Menu bar item?
