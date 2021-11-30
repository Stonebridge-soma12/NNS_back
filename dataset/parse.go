package dataset

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"io/ioutil"
	"mime/multipart"
	"nns_back/cloud"
	"nns_back/log"
	"time"
)

const (
	_csv  = "text/csv"
	_zip  = "application/zip"
	_jpeg = "image/jpeg"
	_png  = "image/png"
)

type ErrUnSupportedContentType struct {
	contentType string
}

func (e ErrUnSupportedContentType) Error() string {
	return fmt.Sprintf("unsupported content type: %s", e.contentType)
}

func IsUnsupportedContentTypeError(err error) bool {
	_, ok := err.(ErrUnSupportedContentType)
	return ok
}

func uploadAsync(storage *cloud.AwsS3Client, file multipart.File, datasetRepo Repository, datasetEntity Dataset) {
	// upload origin file
	mType, err := mimetype.DetectReader(file)
	if err != nil {
		log.Errorw("failed to detect file type",
			"error", err,
			"dataset.id", datasetEntity.ID)
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Errorw("failed to file seek",
			"error", err,
			"dataset.id", datasetEntity.ID)
		return
	}

	var originUrl string
	switch {
	case mType.Is(_csv):
		originUrl, err = storage.UploadFile(file, cloud.WithContentType(_csv), cloud.WithExtension("csv"))
	case mType.Is(_zip):
		originUrl, err = storage.UploadFile(file, cloud.WithContentType(_zip), cloud.WithExtension("zip"))
	default:
		originUrl, err = storage.UploadFile(file)
	}
	if err != nil {
		log.Errorw("failed to upload origin file",
			"error", err,
			"dataset.id", datasetEntity.ID)
		return
	}

	// upload csv file
	// reset file descriptor
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Errorw("failed to file seek",
			"error", err,
			"dataset.id", datasetEntity.ID)
		return
	}

	url, kind, err := save(storage, file)
	if err != nil {
		//if IsUnsupportedContentTypeError(err) {
		//	log.Warn(err)
		//}

		log.Errorw("failed to save file",
			"error", err,
			"dataset.id", datasetEntity.ID)
		return
	}

	datasetEntity, err = datasetRepo.FindByID(datasetEntity.ID)
	if err != nil {
		log.Errorw("failed to find dataset by ID",
			"error", err,
			"dataset.id", datasetEntity.ID)
	}

	switch datasetEntity.Status {
	case UPLOADING:
		datasetEntity.Status = UPLOADED_F
	case UPLOADED_D:
		datasetEntity.Status = EXIST
	default:
		// unexpected
		log.Errorw("dataset status is unexpected",
			"dataset.status", datasetEntity.Status,
			"dataset.id", datasetEntity.ID)
		return
	}
	datasetEntity.OriginURL = sql.NullString{
		Valid:  true,
		String: originUrl,
	}
	datasetEntity.URL = sql.NullString{
		Valid:  true,
		String: url,
	}
	datasetEntity.Kind = kind
	datasetEntity.UpdateTime = time.Now()

	if err := datasetRepo.Update(datasetEntity.ID, datasetEntity); err != nil {
		log.Errorw("failed to update dataset",
			"error", err,
			"dataset.id", datasetEntity.ID)
	}
}

func save(storage *cloud.AwsS3Client, file multipart.File) (string, Kind, error) {
	f, kind, err := parseToDataset(storage, file)
	if err != nil {
		return "", KindUnknown, err
	}

	fBytes, err := io.ReadAll(f)
	if err != nil {
		return "", KindUnknown, err
	}

	url, err := storage.UploadBytes(fBytes, cloud.WithContentType(_csv), cloud.WithExtension("csv"))
	return url, kind, err
}

func parseToDataset(storage *cloud.AwsS3Client, file multipart.File) (io.Reader, Kind, error) {
	mType, err := mimetype.DetectReader(file)
	if err != nil {
		return nil, KindUnknown, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, KindUnknown, err
	}

	switch {
	case mType.Is(_csv):
		return file, KindText, nil

	case mType.Is(_zip):
		return zipToCsv(storage, file)

	default:
		return nil, KindUnknown, ErrUnSupportedContentType{contentType: mType.String()}
	}
}

func zipToCsv(storage *cloud.AwsS3Client, file multipart.File) (io.Reader, Kind, error) {
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, KindUnknown, err
	}

	reader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return nil, KindUnknown, err
	}

	// 압축된 파일 하나하나 읽으면서 jpeg, png 인지 확인. 아니면 에러
	type dataset struct {
		label string
		bytes []byte
	}
	datasetList := make([]dataset, 0)
	var currentDirName string
	for _, zipFile := range reader.File {
		if zipFile.FileInfo().IsDir() {
			currentDirName = zipFile.FileInfo().Name()
			continue
		}

		zipFileBytes, err := readZipFile(zipFile)
		if err != nil {
			return nil, KindUnknown, err
		}

		mType := mimetype.Detect(zipFileBytes)
		switch {
		case mType.Is(_jpeg), mType.Is(_png):
			datasetList = append(datasetList, dataset{
				label: currentDirName,
				bytes: zipFileBytes,
			})

		default:
			return nil, KindUnknown, ErrUnSupportedContentType{contentType: mType.String()}
		}

	}

	buf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buf)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{"url", "label"}); err != nil {
		return nil, KindUnknown, err
	}

	for _, ds := range datasetList {
		url, err := storage.UploadBytes(ds.bytes)
		if err != nil {
			return nil, KindUnknown, err
		}

		log.Debugf("url: %s", url)

		if err := csvWriter.Write([]string{url, ds.label}); err != nil {
			return nil, KindUnknown, err
		}
	}

	return buf, KindImages, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
