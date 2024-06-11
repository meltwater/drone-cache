package s3

// Config is a structure to store S3  backend configuration.
type Config struct {
	// Indicates the files ACL, which should be one,
	// of the following:
	//     private
	//     public-read
	//     public-read-write
	//     authenticated-read
	//     bucket-owner-read
	//     bucket-owner-full-control
	ACL                   string
	Bucket                string
	Encryption            string // if not "", enables server-side encryption. valid values are: AES256, aws:kms.
	Endpoint              string
	Key                   string
	StsEndpoint           string
	AssumeRoleARN         string // if "", do not assume IAM role i.e. use the IAM user.
	AssumeRoleSessionName string
	UserRoleArn           string
	OIDCTokenID           string
	ExternalID            string

	// us-east-1
	// us-west-1
	// us-west-2
	// eu-west-1
	// ap-southeast-1
	// ap-southeast-2
	// ap-northeast-1
	// sa-east-1
	Region string
	Secret string

	PathStyle bool // Use path style instead of domain style. Should be true for minio and false for AWS.
}
