package dataset

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap/zapcore"
	"nns_back/cloud"
	"nns_back/log"
	"os"
	"testing"
)

func TestMnistUploadJob(t *testing.T) {
	log.Init(zapcore.DebugLevel)

	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	//imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")
	datasetBucketName := os.Getenv("DATASET_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	if err != nil {
		t.Errorf("failed to aws load config: %v", err)
		return
	}

	s3Client := s3.NewFromConfig(cfg)
	awsS3Client := &cloud.AwsS3Client{
		Client: s3Client,
		BucketName: datasetBucketName,
	}

	const userId = int64(49) // nns_official
	const fileName = "testdata/trainingSet.zip"
	file, err := os.Open(fileName)
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return
	}

	url, kind, err := save(awsS3Client, file)
	if err != nil {
		t.Errorf("failed to save file: %v", err)
		return
	}

	t.Logf("Success to upload zip file! url: %s, kind: %s", url, kind)
}