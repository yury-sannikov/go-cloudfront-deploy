package main

import (
	"fmt"
	"os"

	ld "github.com/ahl5esoft/golang-underscore"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	flags "github.com/jessevdk/go-flags"
	"github.com/yury-sannikov/go-cloudfront-deploy/cloudfrontTools"
	"github.com/yury-sannikov/go-cloudfront-deploy/fstools"
	"github.com/yury-sannikov/go-cloudfront-deploy/s3tools"
)

type options struct {
	BucketName string `short:"b" long:"bucket_name" description:"S3 bucket name" required:"true"`
	Region     string `short:"r" long:"region" description:"S3 region" required:"true"`
	Path       string `short:"p" long:"path" description:"Path to folder for upload from" required:"false"`
}

type appCommand struct {
	Description string
	Handler     func(*options) error
}

var appCommands = map[string]*appCommand{
	"clean":      &appCommand{Handler: cleanHandler, Description: "Clean bucket"},
	"s3":         &appCommand{Handler: s3DeployHandler, Description: "Copy S3 assets"},
	"cloudfront": &appCommand{Handler: cloudfrontHandler, Description: "Create or Invalidate Cloudfront Distribution"},
}

var opts options

var parser = flags.NewParser(&opts, flags.Default)

func main() {
	args, flagsErr := parser.Parse()
	if flagsErr != nil {
		panic(flagsErr)
	}

	if len(args) == 0 {
		fmt.Println("Please specify action")
		os.Exit(1)
	}

	for _, arg := range args {
		appCommand, err := appCommands[arg]

		fmt.Printf("==>\t%s\n", appCommand.Description)

		if !err {
			fmt.Printf("Unknown action %s\n", args[0])
			os.Exit(2)
		}

		if commandErr := appCommand.Handler(&opts); commandErr != nil {
			fmt.Printf("Error while executing action %v\n", commandErr.Error())
			os.Exit(3)
		}
	}
}

func cleanHandler(options *options) error {
	config := &aws.Config{Region: &options.Region}
	sess := session.New(config)
	svc := s3.New(sess)

	fmt.Printf("Cleaning bucket %s\n", options.BucketName)
	return s3tools.CleanBucket(svc, options.BucketName)
}

func s3DeployHandler(options *options) error {
	config := &aws.Config{Region: &options.Region}
	sess := session.New(config)
	svc := s3.New(sess)

	bucketErr := s3tools.CheckOrCreateBucket(svc, options.BucketName)

	if bucketErr != nil {
		return bucketErr
	}

	files, _ := fstools.ReadFiles(options.Path)

	payloads := ld.Map(files, func(s string, _ int) s3tools.Payload {
		return s3tools.Payload{FilePath: s, S3Service: svc, Bucket: options.BucketName, BaseFolder: options.Path}
	}).([]s3tools.Payload)

	fmt.Printf("Uploading %d files\n", len(payloads))

	dispatcher := s3tools.InitDispatcher(5, 10)

	for _, payload := range payloads {
		work := s3tools.Job{Payload: payload}
		dispatcher.EnqueueJob(work)
	}

	s3tools.DispatcherWaitGroup.Wait()

	fmt.Printf("Done\n")
	return nil
}

func cloudfrontHandler(options *options) error {
	var S3CloudFrontDeployKey = "SiteBuilderCloudfrontID"

	config := &aws.Config{Region: &options.Region}
	sess := session.New(config)
	svc := s3.New(sess)

	tags, err := s3tools.GetBucketTagging(svc, options.BucketName)

	if err != nil {
		return nil
	}
	idx := ld.FindIndex(tags, func(tag *s3.Tag, _ int) bool {
		return *tag.Key == S3CloudFrontDeployKey
	})

	if idx == -1 {
		// Create new Cloudfront deploy as no S3CloudFrontDeployKey has been set
		distributionID, err := cloudfrontTools.Deploy(sess, options.BucketName, options.Region)
		if err != nil {
			return err
		}
		fmt.Printf("New distribution ID has been created: %s\n", *distributionID)
		return s3tools.SetBucketTagging(svc, options.BucketName, S3CloudFrontDeployKey, *distributionID)
	}

	distributionID := tags[idx].Value

	fmt.Printf("Create invalidation request for distribution %s\n", *distributionID)

	return cloudfrontTools.Invalidate(sess, *distributionID)
}
