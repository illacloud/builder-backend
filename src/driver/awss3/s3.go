package awss3

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/illacloud/builder-backend/src/utils/config"
)

type AWSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Region          string
	BucketName      string
	UploadTimeout   time.Duration
	DriveType       int
}

const (
	DRIVE_TYPE_SYSTEM = 1
	DRIVE_TYPE_TEAM   = 2
)
const AWS_DRIVE_URL_ILLA_CLOUD = "illa-cloud-storage.illacloud.com.s3.ap-northeast-1.amazonaws.com"
const AWS_DRIVE_URL_ILLA_CLOUD_TEAM = "illa-cloud-team-storage.illacloud.com.s3.ap-northeast-1.amazonaws.com"
const ILLA_DRIVE_URL_ILLA_CLOUD = "illa-cloud-storage.illacloud.com"
const ILLA_DRIVE_URL_ILLA_CLOUD_TEAM = "illa-cloud-team-storage.illacloud.com"

func NewAWSConfig(endpoint string, accessKeyID string, accessKeySecret string, region string, bucketName string, uploadTimeout int) *AWSConfig {
	timeout := time.Duration(uploadTimeout)
	return &AWSConfig{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Region:          region,
		BucketName:      bucketName,
		UploadTimeout:   timeout,
	}
}

func NewSystemAwsConfigByGlobalConfig(config *config.Config) *AWSConfig {
	return &AWSConfig{
		Endpoint:        config.GetAWSS3Endpoint(),
		AccessKeyID:     config.GetAWSS3AccessKeyID(),
		AccessKeySecret: config.GetAWSS3AccessKeySecret(),
		Region:          config.GetAWSS3Region(),
		BucketName:      config.GetAWSS3SystemBucketName(),
		UploadTimeout:   config.GetAWSS3Timeout(),
		DriveType:       DRIVE_TYPE_SYSTEM,
	}
}

func NewTeamAwsConfigByGlobalConfig(config *config.Config) *AWSConfig {
	return &AWSConfig{
		Endpoint:        config.GetAWSS3Endpoint(),
		AccessKeyID:     config.GetAWSS3AccessKeyID(),
		AccessKeySecret: config.GetAWSS3AccessKeySecret(),
		Region:          config.GetAWSS3Region(),
		BucketName:      config.GetAWSS3TeamBucketName(),
		UploadTimeout:   config.GetAWSS3Timeout(),
		DriveType:       DRIVE_TYPE_TEAM,
	}
}

func CreateAWSSession(awsConfig *AWSConfig) *session.Session {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Endpoint: aws.String(awsConfig.Endpoint),
			Region:   aws.String(awsConfig.Region),
			Credentials: credentials.NewStaticCredentials(
				awsConfig.AccessKeyID,
				awsConfig.AccessKeySecret,
				"",
			),
		},
	))
	return sess
}

func CreateS3Session(awsSession *session.Session) *s3.S3 {
	s3Session := s3.New(awsSession)
	return s3Session
}

type S3Drive struct {
	Instance *s3.S3
	Session  *session.Session
	Config   *AWSConfig
}

func NewS3Drive(awsConfig *AWSConfig) *S3Drive {
	s3Drive := &S3Drive{
		Config: awsConfig,
	}
	s3Drive.Session = CreateAWSSession(awsConfig)
	s3Drive.Instance = CreateS3Session(s3Drive.Session)
	return s3Drive
}

func (s3Drive *S3Drive) GetPreSignedPutURL(fileName string) (string, error) {
	// get put request
	req, _ := s3Drive.Instance.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s3Drive.Config.BucketName),
		Key:    aws.String(fileName), // also you can define path in it.
		ACL:    aws.String("public-read"),
	})

	// sign timeout
	url, errInPresign := req.Presign(s3Drive.Config.UploadTimeout)
	if errInPresign != nil {
		return "", errInPresign
	}

	return url, nil
}
