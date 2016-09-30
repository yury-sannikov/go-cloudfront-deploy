# go-cloudfront-deploy
Deploy static web sites to AWS CloudFront

## Installation
- Install Go using [this](https://golang.org/doc/install]) url
- [Configure](https://golang.org/doc/code.html) Go
- Check path having go/bin folder listed as *export PATH=$PATH:$GOPATH/bin*
- _go install github.com/yury-sannikov/go-cloudfront-deploy_
- You should have _$GOPATH/bin/go-cloudfront-deploy_ executable installed and accessible from command line

## Usage
AWS_ACCESS_KEY_ID and WS_SECRET_ACCESS_KEY environment variables should be set or EC2 role should be enforced in order to make successfull call to AWS API.
If you don't have credentials, go [here](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)

    go-cloudfront-deploy [ACTION LIST] [OPTIONS]
    Action List:
        clean          Clean S3 bucket content before upload any data
        s3             Upload files from specified location to s3 bucket
        cloudfront     Create cloudfront distribution or if any, invalidate it to use new content

    Application Options:
    -b, --bucket_name= S3 bucket name
    -r, --region=      S3 region
    -p, --path=        Path to folder for upload from

    Help Options:
    -h, --help         Show this help message

Action list can contain all of the actions, or some of them, for instance
```
go-cloudfront-deploy clean s3 cloudfront -b site.com -r us-west-1 -p ~/build/site.com
```
Which is equivalent to this:
```
go-cloudfront-deploy clean s3 -b site.com -r us-west-1 -p ~/build/site.com
go-cloudfront-deploy s3 -b site.com -r us-west-1 -p ~/build/site.com
go-cloudfront-deploy cloudfront -b site.com -r us-west-1 -p ~/build/site.com
```

## How it works

### clean
Scan all files in a bucket in batches of 1000 items and delete it.

### s3
- Create new bucket if none exists
- Attach security policy to allow s3:Get for all users
- Enable static web site hosting
- Upload in parallel files from folder
- Set up correct mime types
- Set up cache control headers

### cloudfront
- Check CloudFront distribution using SiteBuilderCloudfrontID tag of correspondent S3 bucket
- If bucket has no distribution, create new one
- If distribution exists, invalidate all objects using /* pattern

