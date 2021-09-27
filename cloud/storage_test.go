package cloud

import (
	"context"
	"crypto/sha256"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestAwsS3Client_Put(t *testing.T) {
	assertions := assert.New(t)

	// Generate test client
	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	assertions.Nil(err)

	s3Client := s3.NewFromConfig(cfg)

	awsS3Client := AwsS3Client{
		Client:     s3Client,
		BucketName: imageBucketName,
	}

	const testFileName = "TestImage.jpg"
	assertions.FileExists(testFileName)
	file, err := os.Open(testFileName)
	assertions.Nil(err)

	objectUrl, err := awsS3Client.Put(file)
	assertions.Nil(err)

	t.Logf("object url : %s", objectUrl)
	resp, err := http.Get(objectUrl)
	assertions.Nil(err)
	defer resp.Body.Close()

	// compare
	_, err = file.Seek(0, io.SeekStart)
	assertions.Nil(err)

	baseFileBytes, err := ioutil.ReadAll(file)
	assertions.Nil(err)

	responseFileBytes, err := ioutil.ReadAll(resp.Body)
	assertions.Nil(err)

	baseChecksum := sha256.Sum256(baseFileBytes)
	responseChecksum := sha256.Sum256(responseFileBytes)
	assertions.Equal(baseChecksum, responseChecksum)
}

func TestAwsS3Client_PutBytes(t *testing.T) {
	assertions := assert.New(t)

	// Generate test client
	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	assertions.Nil(err)

	s3Client := s3.NewFromConfig(cfg)

	awsS3Client := AwsS3Client{
		Client:     s3Client,
		BucketName: imageBucketName,
	}

	const testFileName = "TestImage.jpg"
	assertions.FileExists(testFileName)
	file, err := os.Open(testFileName)
	assertions.Nil(err)

	fileBytes, err := io.ReadAll(file)
	assertions.Nil(err)

	objectUrl, err := awsS3Client.PutBytes(fileBytes)
	assertions.Nil(err)

	t.Logf("object url : %s", objectUrl)
	resp, err := http.Get(objectUrl)
	assertions.Nil(err)
	defer resp.Body.Close()

	// compare
	_, err = file.Seek(0, io.SeekStart)
	assertions.Nil(err)

	baseFileBytes, err := ioutil.ReadAll(file)
	assertions.Nil(err)

	responseFileBytes, err := ioutil.ReadAll(resp.Body)
	assertions.Nil(err)

	baseChecksum := sha256.Sum256(baseFileBytes)
	responseChecksum := sha256.Sum256(responseFileBytes)
	assertions.Equal(baseChecksum, responseChecksum)
}