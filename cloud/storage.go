package cloud

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"io"
	"mime/multipart"
	"strings"
	"time"
)

type AwsS3Client struct {
	Client     *s3.Client
	BucketName string
}

type Option interface {
	apply(input *s3.PutObjectInput)
}

type optionFunc func(input *s3.PutObjectInput)

func (f optionFunc) apply(input *s3.PutObjectInput) {
	f(input)
}

func WithContentType(contentType string) Option {
	return optionFunc(func(input *s3.PutObjectInput) {
		input.ContentType = aws.String(contentType)
	})
}

func WithExtension(extension string) Option {
	return optionFunc(func(input *s3.PutObjectInput) {
		filename := *(input.Key)
		splitedFilename := strings.Split(filename, ".")
		splitedFilename[len(splitedFilename) - 1] = extension
		input.Key = aws.String(strings.Join(splitedFilename, "."))
	})
}

func (c *AwsS3Client) Put(file multipart.File, options ...Option) (url string, err error) {
	mType, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	filename := generateFileName(mType.Extension())

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.BucketName),
		Key: aws.String(filename),
		Body: file,
		ContentType: aws.String(mType.String()),
		ACL: types.ObjectCannedACLPublicRead,
	}

	for _, option := range options {
		option.apply(input)
	}

	if _, err := c.Client.PutObject(context.TODO(), input); err != nil {
		return "", err
	}

	return getS3ObjectUrl(c.BucketName, filename), nil
}

func (c *AwsS3Client) PutBytes(file []byte, options ...Option) (url string, err error) {
	mType := mimetype.Detect(file)

	filename := generateFileName(mType.Extension())

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.BucketName),
		Key: aws.String(filename),
		Body: bytes.NewReader(file),
		ContentType: aws.String(mType.String()),
		ACL: types.ObjectCannedACLPublicRead,
	}

	for _, option := range options {
		option.apply(input)
	}

	if _, err := c.Client.PutObject(context.TODO(), input); err != nil {
		return "", err
	}

	return getS3ObjectUrl(c.BucketName, *input.Key), nil
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