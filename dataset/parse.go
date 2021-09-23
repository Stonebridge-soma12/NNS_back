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

	return storage.PutBytes(fBytes, cloud.WithContentType(_csv), cloud.WithExtension("csv"))
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

	zipFileBytesList := make([][]byte, 0, len(reader.File))
	for _, zipFile := range reader.File {
		zipFileBytes, err := readZipFile(zipFile)
		if err != nil {
			return nil, err
		}

		mType := mimetype.Detect(zipFileBytes)
		switch {
		case mType.Is(_jpeg), mType.Is(_png):
			zipFileBytesList = append(zipFileBytesList, zipFileBytes)

		default:
			return nil, ErrUnSupportedContentType{contentType: mType.String()}
		}

	}

	buf := new(bytes.Buffer)
	csvWriter := csv.NewWriter(buf)
	defer csvWriter.Flush()

	if err := csvWriter.Write([]string{"url"}); err != nil {
		return nil, err
	}

	for _, zipFileBytes := range zipFileBytesList {
		url, err := storage.PutBytes(zipFileBytes)
		if err != nil {
			return nil, err
		}

		if err := csvWriter.Write([]string{url}); err != nil {
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