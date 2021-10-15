package dataset

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"io"
	"io/ioutil"
	"mime/multipart"
	"nns_back/cloud"
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

func save(storage *cloud.AwsS3Client, file multipart.File) (string, error) {
	f, err := parseToDataset(storage, file)
	if err != nil {
		return "", err
	}

	fBytes, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return storage.UploadBytes(fBytes, cloud.WithContentType(_csv), cloud.WithExtension("csv"))
}

func parseToDataset(storage *cloud.AwsS3Client, file multipart.File) (io.Reader, error) {
	mType, err := mimetype.DetectReader(file)
	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	if mType.Is(_csv) {
		return file, nil
	}

	switch {
	case mType.Is(_csv):
		return file, nil

	case mType.Is(_zip):
		return zipToCsv(storage, file)

	default:
		return nil, ErrUnSupportedContentType{contentType: mType.String()}
	}
}

func zipToCsv(storage *cloud.AwsS3Client, file multipart.File) (io.Reader, error) {
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return nil, err
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
			return nil, err
		}

		mType := mimetype.Detect(zipFileBytes)
		switch {
		case mType.Is(_jpeg), mType.Is(_png):
			datasetList = append(datasetList, dataset{
				label: currentDirName,
				bytes: zipFileBytes,
			})

		default:
			return nil, ErrUnSupportedContentType{contentType: mType.String()}
		}

	}

	buf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buf)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{"url", "label"}); err != nil {
		return nil, err
	}

	for _, ds := range datasetList {
		url, err := storage.UploadBytes(ds.bytes)
		if err != nil {
			return nil, err
		}

		if err := csvWriter.Write([]string{url, ds.label}); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}
