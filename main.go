package main

import (
	"flag"
	"fmt"

	ld "github.com/ahl5esoft/golang-underscore"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/yury-sannikov/go-cloudfront-deploy/fstools"
	"github.com/yury-sannikov/go-cloudfront-deploy/s3tools"
)

// 1 - Add site under hosted zone
// .. domain nanme
// .. description
// ==> JSON with

var (
	bucketName string
	region     string
	baseFolder string
	dropBucket bool
)

func init() {
	const (
		defaultBucketName = ""
		usageBucket       = "Amazon S3 bucket name to deploy"
		defaultRegionName = "us-west-1"
		usageRegion       = "Amazon S3 region"
		defaultFolder     = "."
		usageFolder       = "Path to assets folder"
	)
	flag.StringVar(&bucketName, "bucket_name", defaultBucketName, usageBucket)
	flag.StringVar(&bucketName, "b", defaultBucketName, usageBucket+" (shorthand)")
	flag.StringVar(&region, "region", defaultRegionName, usageRegion)
	flag.StringVar(&region, "r", defaultRegionName, usageRegion+" (shorthand)")
	flag.StringVar(&baseFolder, "base", defaultFolder, usageFolder)
	flag.StringVar(&baseFolder, "f", defaultFolder, usageFolder+" (shorthand)")
	flag.BoolVar(&dropBucket, "drop", false, "Drop S3 bucket initially")
}

func main() {
	flag.Parse()
	if bucketName == "" {
		flag.PrintDefaults()
		return
	}
	fmt.Printf("Region: %s, Bucket Name: %s\n", region, bucketName)

	config := &aws.Config{Region: &region}
	sess := session.New(config)
	svc := s3.New(sess)

	if dropBucket {
		fmt.Printf("Dropping bucket %s\n", bucketName)
		s3tools.DropBucket(svc, bucketName)
	}

	bucketErr := s3tools.CheckOrCreateBucket(svc, bucketName)

	if bucketErr != nil {
		return
	}

	files, _ := fstools.ReadFiles(baseFolder)

	payloads := ld.Map(files, func(s string, _ int) s3tools.Payload {
		return s3tools.Payload{FilePath: s, S3Service: svc, Bucket: bucketName, BaseFolder: baseFolder}
	}).([]s3tools.Payload)

	fmt.Printf("Uploading %d files\n", len(payloads))

	dispatcher := s3tools.InitDispatcher(5, 10)

	for _, payload := range payloads {
		work := s3tools.Job{Payload: payload}
		dispatcher.EnqueueJob(work)
	}

	s3tools.DispatcherWaitGroup.Wait()

	fmt.Printf("Done\n")
}
