package cloud

import (
	"context"
	"crypto/sha256"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"io/ioutil"
	"net/http"
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
	resp, err := http.Get(objectUrl)
	if err != nil {
		t.Errorf("failed to get image from object url: %+v", err)
	}
	defer resp.Body.Close()

	// compare
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Errorf("failed to file seek: %+v", err)
	}

	baseFileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("failed to read base file: %+v", err)
	}

	responseFileBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read response file: %+v", err)
	}

	baseChecksum := sha256.Sum256(baseFileBytes)
	responseChecksum := sha256.Sum256(responseFileBytes)
	if baseChecksum != responseChecksum {
		t.Logf("base checksum : %x", baseChecksum)
		t.Logf("response checksum : %x", responseChecksum)
		t.Fatal("image not equal")
	}
}
