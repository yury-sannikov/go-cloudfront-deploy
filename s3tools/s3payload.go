package s3tools

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Payload is a type to handle worker operations
type Payload struct {
	Bucket    string
	FilePath  string
	S3Service *s3.S3
}

func (payload Payload) process() error {
	file, ferr := os.Open(payload.FilePath)
	if ferr != nil {
		log.Fatal("Failed to open file", ferr)
		return ferr
	}

	uploader := s3manager.NewUploaderWithClient(payload.S3Service)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Body:   file,
		Bucket: &payload.Bucket,
		Key:    &payload.FilePath,
	})

	if err != nil {
		fmt.Printf("Error while processing %s, %s\n", payload.FilePath, err)
	}
	return err
}
