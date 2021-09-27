package dataset

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/stretchr/testify/assert"
	"nns_back/cloud"
	"os"
	"testing"
)

func Test_validateMimetype(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "csv",
			path: "testdata/csv.csv",
			want: _csv,
		},
		{
			name: "zip",
			path: "testdata/zip.zip",
			want: _zip,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			file, err := os.Open(tt.path)
			assert.Nil(err)

			mType, err := mimetype.DetectReader(file)
			assert.Nil(err)

			assert.True(mType.Is(tt.want))
			assert.Equal(mType.String(), tt.want)
		})
	}
}

func Test_save(t *testing.T) {
	assertions := assert.New(t)

	// Generate test client
	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	imageBucketName := os.Getenv("DATASET_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	assertions.Nil(err)

	s3Client := s3.NewFromConfig(cfg)

	awsS3Client := cloud.AwsS3Client{
		Client:     s3Client,
		BucketName: imageBucketName,
	}

	tests := []struct {
		name    string
		path    string
		wanterr bool
	}{
		{
			name:    "save csv.csv",
			path:    "testdata/csv.csv",
			wanterr: false,
		},
		{
			name:    "save zip.zip (image zip)",
			path:    "testdata/zip.zip",
			wanterr: false,
		},
		{
			name:    "can not save unsupported content type: gz",
			path:    "testdata/t10k-images-idx3-ubyte.gz",
			wanterr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.path)
			assertions.Nil(err)

			url, err := save(&awsS3Client, f)
			if (err != nil) != tt.wanterr {
				t.Errorf("save() error = %v, wanterr %v", err, tt.wanterr)
				return
			}
			if err != nil {
				t.Logf("error: %v", err)
			} else {
				t.Logf("object url : %s", url)
			}
		})
	}
}
