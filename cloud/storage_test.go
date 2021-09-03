package cloud

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"os"
	"testing"
)

func TestAwsS3Client_Put(t *testing.T) {
	// Generate test client
	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)

	if err != nil {
		t.Errorf("failed to load aws config: %+v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	awsS3Client := AwsS3Client{
		Client:     s3Client,
		BucketName: imageBucketName,
	}

	file, err := os.Open("TestImage.jpg")
	if err != nil {
		t.Errorf("failed to open file: %+v", err)
	}

	objectUrl, err := awsS3Client.Put(file)
	if err != nil {
		t.Errorf("failed to put object: %+v", err)
	}

	t.Logf("object url : %s", objectUrl)
}
