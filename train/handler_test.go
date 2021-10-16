package train

import (
	"github.com/stretchr/testify/assert"
	"nns_back/model"
	"nns_back/util"
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

func Test_getDatasetConfigId(t *testing.T) {
	project := model.Project{
		Config: util.NullJson{
			Valid: true,
			Json: []byte(`{
    "optimizer_name": "Adam",
    "optimizer_config": {
        "learning_rate": 0.001,
        "beta_1": 0.9,
        "beta_2": 0.999,
        "epsilon": 1e-07,
        "amsgrad": false
    },
    "loss": "binary_crossentropy",
    "metrics": [
        "accuracy"
    ],
    "batch_size": 32,
    "epochs": 10,
    "early_stop": {
        "usage": true,
        "monitor": "loss",
        "patience": 2
    },
    "learning_rate_reduction": {
        "usage": true,
        "monitor": "val_accuracy",
        "patience": 2,
        "factor": 0.25,
        "min_lr": 0.0000003
    },
    "dataset_config": {
        "valid": true,
        "id": 1
    }
}`),
		},
	}

	dscId, err := getDatasetConfigId(project)
	assert.NoError(t, err)
	assert.EqualValues(t, 1, dscId)
}
