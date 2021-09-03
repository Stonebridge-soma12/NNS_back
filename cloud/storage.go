package cloud

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type AwsS3Client struct {
	Client     *s3.Client
	BucketName string
}

func (c *AwsS3Client) Put(file multipart.File) (url string, err error) {
	contentType, err := getFileContentType(file)
	if err != nil {
		return "", err
	}

	fileExtension := "." + strings.Split(contentType, "/")[1]
	filename := generateFileName(fileExtension)

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.BucketName),
		Key: aws.String(filename),
		Body: file,
		ContentType: aws.String(contentType),
		ACL: types.ObjectCannedACLPublicRead,
	}

	if _, err := c.Client.PutObject(context.TODO(), input); err != nil {
		return "", err
	}

	return getS3ObjectUrl(c.BucketName, filename), nil
}

func getFileContentType(seeker io.ReadSeeker) (string, error) {
	// At most the first 512 bytes of data are used:
	// https://golang.org/src/net/http/sniff.go?s=646:688#L11
	buff := make([]byte, 512)

	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	n, err := seeker.Read(buff)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Slice to remove fill-up zero values which cause a wrong content type detection in the next step
	buff = buff[:n]

	// Reset the read pointer if necessary
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return http.DetectContentType(buff), nil
}

func generateFileName(addLast ...string) string {
	const _fileNameTimeLayout = "2006/01/02/"
	fileName := time.Now().Format(_fileNameTimeLayout) + uuid.NewString()
	for _, v := range addLast {
		fileName += v
	}

	return fileName
}

func getS3ObjectUrl(bucketName, fileName string) string {
	const _s3ObjectUrlPrefix = "https://s3.ap-northeast-2.amazonaws.com/"
	return _s3ObjectUrlPrefix + bucketName + "/" + fileName
}