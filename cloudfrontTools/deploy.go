package cloudfrontTools

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

//Deploy creates cloudfront tagged distribution
func Deploy(session *session.Session, bucketName string, region string) (*string, error) {

	svc := cloudfront.New(session)

	params := &cloudfront.CreateDistributionWithTagsInput{
		DistributionConfigWithTags: &cloudfront.DistributionConfigWithTags{
			DistributionConfig: createDistributionConfig(bucketName, region),
			Tags: &cloudfront.Tags{
				Items: []*cloudfront.Tag{
					{
						Key:   aws.String("bucketName"),
						Value: &bucketName,
					},
				},
			},
		},
	}

	resp, err := svc.CreateDistributionWithTags(params)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Printf("Successfully created distribution id=%s\n", *resp.Distribution.Id)

	return resp.Distribution.Id, err
}

func createDistributionConfig(bucketName string, region string) *cloudfront.DistributionConfig {
	callerRef := fmt.Sprintf("%v", time.Now().Unix())
	comment := fmt.Sprintf("Deployed via go-cloudfront-deploy at %s", time.Now().Format(time.UnixDate))

	domainName := fmt.Sprintf("%s.s3-website-%s.amazonaws.com", bucketName, region)
	targetOriginID := fmt.Sprintf("sitebuilder-%s", domainName)
	wwwAlias := fmt.Sprintf("www.%s", bucketName)

	return &cloudfront.DistributionConfig{ // Required
		CallerReference: &callerRef,
		Comment:         &comment,
		DefaultCacheBehavior: &cloudfront.DefaultCacheBehavior{
			ForwardedValues: &cloudfront.ForwardedValues{
				Cookies: &cloudfront.CookiePreference{
					Forward: aws.String("none"),
					WhitelistedNames: &cloudfront.CookieNames{
						Quantity: aws.Int64(0),
						Items:    []*string{},
					},
				},
				QueryString: aws.Bool(false),
				Headers: &cloudfront.Headers{
					Quantity: aws.Int64(0),
					Items:    []*string{},
				},
				QueryStringCacheKeys: &cloudfront.QueryStringCacheKeys{
					Quantity: aws.Int64(0),
					Items:    []*string{},
				},
			},
			MinTTL:         aws.Int64(0),
			TargetOriginId: &targetOriginID,
			TrustedSigners: &cloudfront.TrustedSigners{
				Enabled:  aws.Bool(false),
				Quantity: aws.Int64(0),
				Items:    []*string{},
			},
			ViewerProtocolPolicy: aws.String("allow-all"),
			AllowedMethods: &cloudfront.AllowedMethods{
				Items: []*string{
					aws.String("GET"),
					aws.String("HEAD"),
				},
				Quantity: aws.Int64(2),
				CachedMethods: &cloudfront.CachedMethods{
					Items: []*string{ // Required
						aws.String("GET"),
						aws.String("HEAD"),
					},
					Quantity: aws.Int64(2),
				},
			},
			Compress:        aws.Bool(true),
			DefaultTTL:      aws.Int64(86400),
			MaxTTL:          aws.Int64(31536000),
			SmoothStreaming: aws.Bool(false),
		},
		Enabled: aws.Bool(true),
		Origins: &cloudfront.Origins{
			Quantity: aws.Int64(1),
			Items: []*cloudfront.Origin{
				{
					DomainName: &domainName,
					Id:         &targetOriginID,
					CustomHeaders: &cloudfront.CustomHeaders{
						Quantity: aws.Int64(0),
						Items:    []*cloudfront.OriginCustomHeader{},
					},
					CustomOriginConfig: &cloudfront.CustomOriginConfig{
						HTTPPort:             aws.Int64(80),
						HTTPSPort:            aws.Int64(443),
						OriginProtocolPolicy: aws.String("http-only"),
						OriginSslProtocols: &cloudfront.OriginSslProtocols{
							Items: []*string{
								aws.String("TLSv1.1"),
								aws.String("TLSv1.2"),
							},
							Quantity: aws.Int64(2),
						},
					},
					OriginPath: aws.String(""),
					// S3OriginConfig: &cloudfront.S3OriginConfig{
					// 	OriginAccessIdentity: aws.String("string"), // Required
					// },
				},
			},
		},
		Aliases: &cloudfront.Aliases{
			Quantity: aws.Int64(2),
			Items: []*string{
				aws.String(bucketName),
				aws.String(wwwAlias),
			},
		},
		CacheBehaviors: &cloudfront.CacheBehaviors{
			Quantity: aws.Int64(0), // Required
			//Items: []*cloudfront.CacheBehavior{
			//{
			//     // Required
			// 	ForwardedValues: &cloudfront.ForwardedValues{ // Required
			// 		Cookies: &cloudfront.CookiePreference{ // Required
			// 			Forward: aws.String("ItemSelection"), // Required
			// 			WhitelistedNames: &cloudfront.CookieNames{
			// 				Quantity: aws.Int64(1), // Required
			// 				Items: []*string{
			// 					aws.String("string"), // Required
			// 					// More values...
			// 				},
			// 			},
			// 		},
			// 		QueryString: aws.Bool(true), // Required
			// 		Headers: &cloudfront.Headers{
			// 			Quantity: aws.Int64(1), // Required
			// 			Items: []*string{
			// 				aws.String("string"), // Required
			// 				// More values...
			// 			},
			// 		},
			// 		QueryStringCacheKeys: &cloudfront.QueryStringCacheKeys{
			// 			Quantity: aws.Int64(1), // Required
			// 			Items: []*string{
			// 				aws.String("string"), // Required
			// 				// More values...
			// 			},
			// 		},
			// 	},
			// 	MinTTL:         aws.Int64(1),         // Required
			// 	PathPattern:    aws.String("string"), // Required
			// 	TargetOriginId: aws.String("string"), // Required
			// 	TrustedSigners: &cloudfront.TrustedSigners{ // Required
			// 		Enabled:  aws.Bool(true), // Required
			// 		Quantity: aws.Int64(1),   // Required
			// 		Items: []*string{
			// 			aws.String("string"), // Required
			// 			// More values...
			// 		},
			// 	},
			// 	ViewerProtocolPolicy: aws.String("ViewerProtocolPolicy"), // Required
			// 	AllowedMethods: &cloudfront.AllowedMethods{
			// 		Items: []*string{ // Required
			// 			aws.String("Method"), // Required
			// 			// More values...
			// 		},
			// 		Quantity: aws.Int64(1), // Required
			// 		CachedMethods: &cloudfront.CachedMethods{
			// 			Items: []*string{ // Required
			// 				aws.String("Method"), // Required
			// 				// More values...
			// 			},
			// 			Quantity: aws.Int64(1), // Required
			// 		},
			// 	},
			// 	Compress:        aws.Bool(true),
			// 	DefaultTTL:      aws.Int64(1),
			// 	MaxTTL:          aws.Int64(1),
			// 	SmoothStreaming: aws.Bool(true),
			//},
			// More values...
			//},
		},
		CustomErrorResponses: &cloudfront.CustomErrorResponses{
			Quantity: aws.Int64(0),
			// Items: []*cloudfront.CustomErrorResponse{
			// 	{
			// 	// ErrorCode:          aws.Int64(1), // Required
			// 	// ErrorCachingMinTTL: aws.Int64(1),
			// 	// ResponseCode:       aws.String("string"),
			// 	// ResponsePagePath:   aws.String("string"),
			// 	},
			// 	// More values...
			// },
		},
		DefaultRootObject: aws.String("index.html"),
		HttpVersion:       aws.String(cloudfront.HttpVersionHttp2),
		Logging: &cloudfront.LoggingConfig{
			Bucket:         aws.String(""),
			Enabled:        aws.Bool(false),
			IncludeCookies: aws.Bool(false),
			Prefix:         aws.String(""),
		},
		PriceClass: aws.String(cloudfront.PriceClassPriceClassAll),
		Restrictions: &cloudfront.Restrictions{
			GeoRestriction: &cloudfront.GeoRestriction{
				Quantity:        aws.Int64(0),
				RestrictionType: aws.String("none"),
				Items:           []*string{},
			},
		},
		// TODO:
		// ViewerCertificate: &cloudfront.ViewerCertificate{
		// 	ACMCertificateArn:            aws.String("string"),
		// 	Certificate:                  aws.String("string"),
		// 	CertificateSource:            aws.String("CertificateSource"),
		// 	CloudFrontDefaultCertificate: aws.Bool(true),
		// 	IAMCertificateId:             aws.String("string"),
		// 	MinimumProtocolVersion:       aws.String("MinimumProtocolVersion"),
		// 	SSLSupportMethod:             aws.String("SSLSupportMethod"),
		// },
		WebACLId: aws.String(""),
	}
}

//Invalidate entire cloudfront distribution
func Invalidate(session *session.Session, distributionID string) error {
	svc := cloudfront.New(session)

	callerRef := fmt.Sprintf("%v", time.Now().Unix())

	params := &cloudfront.CreateInvalidationInput{
		DistributionId: &distributionID,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &callerRef,
			Paths: &cloudfront.Paths{
				Quantity: aws.Int64(1),
				Items: []*string{
					aws.String("/*"),
				},
			},
		},
	}
	resp, err := svc.CreateInvalidation(params)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("Invalidation has been created %s\n", *resp.Location)
	return nil
}
