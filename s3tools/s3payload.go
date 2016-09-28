package s3tools

import (
	"fmt"
	"log"
	"mime"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Payload is a type to handle worker operations
type Payload struct {
	Bucket     string
	FilePath   string
	S3Service  *s3.S3
	BaseFolder string
}

func (payload Payload) process() error {
	file, ferr := os.Open(payload.FilePath)
	if ferr != nil {
		log.Fatal("Failed to open file", ferr)
		return ferr
	}
	rel, errRel := filepath.Rel(payload.BaseFolder, payload.FilePath)

	if errRel != nil {
		return errRel
	}

	uploader := s3manager.NewUploaderWithClient(payload.S3Service)

	uploadInput := s3manager.UploadInput{
		Body:   file,
		Bucket: &payload.Bucket,
		Key:    &rel,
	}
	updateMetadata(rel, &uploadInput)

	_, err := uploader.Upload(&uploadInput)

	if err != nil {
		fmt.Printf("Error while processing %s, %s\n", payload.FilePath, err)
	}
	return err
}

func updateMetadata(name string, input *s3manager.UploadInput) {
	var DefaultCacheControl = "max-age=31536000,public"
	var HTMLCacheControl = "max-age=86400,public"

	input.CacheControl = &DefaultCacheControl

	fileExt := path.Ext(name)

	if fileExt == ".html" {
		input.CacheControl = &HTMLCacheControl
	}

	mimeType := mime.TypeByExtension(fileExt)

	// Fallback to default mime type
	if len(mimeType) == 0 {
		mimeType = "application/octet-stream"
	}

	input.ContentType = &mimeType
}
