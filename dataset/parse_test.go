package dataset

import (
	"github.com/gabriel-vasile/mimetype"
	"github.com/stretchr/testify/assert"
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