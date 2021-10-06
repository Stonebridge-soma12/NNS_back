package train

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_testify연습(t *testing.T) {
	assert := assert.New(t)

	repo := MockTrainRepository{}
	repo.On("FindNextTrainNo", int64(1)).Return(int64(1), nil)
	repo.On("FindNextTrainNo", int64(2)).Return(int64(1), nil)
	defer repo.AssertExpectations(t)
	defer repo.AssertNumberOfCalls(t, "FindNextTrainNo", 2)

	nextTrainNo, err := repo.FindNextTrainNo(1)
	assert.NoError(err)
	assert.EqualValues(1, nextTrainNo)

	nextTrainNo, err = repo.FindNextTrainNo(2)
	assert.NoError(err)
	assert.EqualValues(1, nextTrainNo)
}