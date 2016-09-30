package s3tools

import (
	"fmt"
	"log"
	"reflect"

	ld "github.com/ahl5esoft/golang-underscore"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

const _NotFoundError = "NotFound"
const _NoWebConfig = "NoSuchWebsiteConfiguration"

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
	return checkBucketWebsiteConfiguration(service, bucketName)
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

	policyErr := addBucketPolicy(service, bucketName)

	if policyErr != nil {
		return policyErr
	}

	return checkBucketWebsiteConfiguration(service, bucketName)
}

func checkBucketWebsiteConfiguration(service *s3.S3, bucketName string) error {
	_, err := service.GetBucketWebsite(&s3.GetBucketWebsiteInput{
		Bucket: &bucketName,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == _NoWebConfig {
				return createBucketWebsiteConfiguration(service, bucketName)
			}
		}
	}
	if err != nil {
		fmt.Printf("checkBucketWebsiteConfiguration error %s\n", err)
	}
	return err
}

func createBucketWebsiteConfiguration(service *s3.S3, bucketName string) error {
	indexSuffix := "index.html"
	_, err := service.PutBucketWebsite(&s3.PutBucketWebsiteInput{
		Bucket: &bucketName,
		WebsiteConfiguration: &s3.WebsiteConfiguration{
			IndexDocument: &s3.IndexDocument{
				Suffix: &indexSuffix,
			},
		},
	})
	if err == nil {
		fmt.Printf("Static Website Configuration updated\n")
	}
	return err
}

func addBucketPolicy(service *s3.S3, bucketName string) error {
	webAcl :=
		`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "AddPerm",
                "Effect": "Allow",
                "Principal": "*",
                "Action": "s3:GetObject",
                "Resource": "arn:aws:s3:::%s/*"
            }
        ]
    }`
	policy := fmt.Sprintf(webAcl, bucketName)
	_, perr := service.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: &bucketName,
		Policy: &policy,
	})

	if perr != nil {
		fmt.Printf("Unable to set up ACL policy for %s: %s\n", bucketName, perr)
	}
	return perr
}

//DropBucket drop bucket
func DropBucket(service *s3.S3, bucketName string) error {
	err := service.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {

		fmt.Printf("Deleting chunk of %d objects\n", len(p.Contents))

		if len(p.Contents) == 0 {
			return false
		}

		objs := ld.Map(p.Contents, func(o *s3.Object, _ int) *s3.ObjectIdentifier {
			return &s3.ObjectIdentifier{
				Key: o.Key,
			}
		}).([]*s3.ObjectIdentifier)

		_, err := service.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: &bucketName,
			Delete: &s3.Delete{
				Objects: objs,
			},
		})

		if err != nil {
			fmt.Printf("failed to delete objects %s\n", err)
			return false
		}

		return true
	})

	if err != nil {
		fmt.Printf("failed to list objects %s\n", err)
	} else {
		fmt.Println("Done")
	}
	return err
}

//GetBucketTagging return an array of associated S3 tags
func GetBucketTagging(service *s3.S3, bucketName string) ([]*s3.Tag, error) {
	params := &s3.GetBucketTaggingInput{
		Bucket: &bucketName,
	}
	resp, err := service.GetBucketTagging(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "NoSuchTagSet" {
				var tagSet []*s3.Tag
				return tagSet, nil
			}
		}
		fmt.Println(err.Error())
		return nil, err
	}
	return resp.TagSet, nil
}

//SetBucketTagging replace tag with specified key/value
func SetBucketTagging(service *s3.S3, bucketName string, tagKey string, tagValue string) error {
	params := &s3.PutBucketTaggingInput{
		Bucket: &bucketName,
		Tagging: &s3.Tagging{
			TagSet: []*s3.Tag{
				{
					Key:   &tagKey,
					Value: &tagValue,
				},
			},
		},
	}

	_, err := service.PutBucketTagging(params)

	return err
}
