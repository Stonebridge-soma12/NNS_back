package service

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"io"
	"net/http"
	"nns_back/model"
	"nns_back/util"
	"os"
	"strings"
	"time"
)

const _uploadImageFormFileKey = "image"

func (e Env) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	// maximum upload of 10 MB files
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		e.Logger.Errorw("failed to specifies a maximum file size.",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	file, header, err := r.FormFile(_uploadImageFormFileKey)
	if err != nil {
		e.Logger.Errorw("failed to retrieve form file",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}
	defer file.Close()

	e.Logger.Debugw("success to retrieve form file",
		"file name", header.Filename,
		"file size", header.Size,
		"MIME header", header.Header)

	url, err := UploadImage(file, header.Header.Get("Content-Type"))
	if err != nil {
		e.Logger.Errorw("failed to upload image to s3",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	e.Logger.Debugw(url)

	userId, ok := r.Context().Value("userId").(int64)
	if !ok {
		e.Logger.Errorw("failed to conversion interface to int64",
			"error code", util.ErrInternalServerError,
			"context value", r.Context().Value("userId"))
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}

	img := model.NewImage(userId, url)
	if img.Id, err = img.Insert(e.DB); err != nil {
		e.Logger.Errorw("failed to insert image",
			"error code", util.ErrInternalServerError,
			"error", err)
		util.WriteError(w, http.StatusInternalServerError, util.ErrInternalServerError)
		return
	}


	util.WriteJson(w, http.StatusCreated, util.ResponseBody{
		"id": img.Id,
		"url": img.Url,
	})
}

func UploadImage(file io.Reader, contentType string) (string, error) {
	// put file to S3
	awsAccessKeyId := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsSessionToken := os.Getenv("AWS_SESSION_TOKEN")
	imageBucketName := os.Getenv("IMAGE_BUCKET_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyId, awsSecretAccessKey, awsSessionToken)),
		config.WithRegion("ap-northeast-2"),
	)
	if err != nil {
		return "", err
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)
	bucket := aws.String(imageBucketName)
	filename := aws.String(time.Now().Format("2006/01/02/") +
		uuid.NewString() +
		"." +
		strings.Split(contentType, "/")[1])

	input := &s3.PutObjectInput{
		Bucket:      bucket,
		Key:         filename,
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	}

	if _, err := client.PutObject(context.TODO(), input); err != nil {
		return "", err
	}

	const urlPrefix = "https://s3.ap-northeast-2.amazonaws.com/"
	return urlPrefix + imageBucketName + "/" + *filename, nil
}
