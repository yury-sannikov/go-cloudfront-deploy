package s3tools

import (
	"fmt"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

const _NotFoundError = "NotFound"

// CheckOrCreateBucket check if specified bucket exists, if not, creates new one
func CheckOrCreateBucket(service *s3.S3, bucketName string) error {
	fmt.Printf("Checking bucket name %s\n", bucketName)

	_, err := service.HeadBucket(&s3.HeadBucketInput{Bucket: &bucketName})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == _NotFoundError {
				fmt.Printf("Bucket '%s' does not exists\n", bucketName)
				return createBucket(service, bucketName)
			}
		}
		log.Printf("Error %s while checking bucket %s: %+v\n", reflect.TypeOf(err), bucketName, err)
		return err
	}
	fmt.Printf("Bucket '%s' exists\n", bucketName)
	return nil
}

func createBucket(service *s3.S3, bucketName string) error {
	result, err := service.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucketName,
	})

	if err != nil {
		log.Println("Failed to create bucket.", err)
		return err
	}

	if err = service.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &bucketName}); err != nil {
		log.Printf("Failed to wait for bucket to exist %s, %s\n", bucketName, err)
		return err
	}

	log.Printf("Successfully created bucket %s\n", result)
	return nil
}
