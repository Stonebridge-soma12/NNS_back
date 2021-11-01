package model

import (
	"database/sql"
	"encoding/json"
	"nns_back/util"
	"time"
)

type Project struct {
	Id          int64          `db:"id"`
	ShareKey    sql.NullString `db:"share_key"`
	UserId      int64          `db:"user_id"`
	ProjectNo   int            `db:"project_no"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Config      util.NullJson  `db:"config"`
	Content     util.NullJson  `db:"content"`
	Status      util.Status    `db:"status"`
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
}

func NewProject(userId int64, projectNo int, name, description string) Project {
	return Project{
		ShareKey:    sql.NullString{Valid: false},
		UserId:      userId,
		ProjectNo:   projectNo,
		Name:        name,
		Description: description,
		Config:      DefaultConfig(),
		Content:     DefaultContent(),
		Status:      util.StatusEXIST,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
}

const (
	_defaultProjectConfig = `{
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
        "valid": false,
        "id": 0
    }
}`

	_defaultProjectContent = `{
    "output": "inputnode_1",
    "flowState": {
        "elements": [
            {
                "id": "node_default_input_node_auto_created",
                "type": "Layer",
                "position": {
                    "x": 62.140625,
                    "y": 86
                },
                "data": {
                    "label": "InputNode_1",
                    "category": "Layer",
                    "param": {
                        "shape": []
                    },
                    "type": "Input"
                }
            }
        ],
        "position": [
            100,
            100
        ],
        "zoom": 1
    },
    "input": "inputnode_1",
    "layers": [
        {
            "category": "Layer",
            "type": "Input",
            "name": "inputnode_1",
            "id": "node_default_input_node_auto_created",
            "input": [],
            "output": [],
            "param": {
                "shape": []
            }
        }
    ]
}`
)

func DefaultConfig() util.NullJson {
	return util.NullJson{
		Json:  json.RawMessage(_defaultProjectConfig),
		Valid: true,
	}
}

func DefaultContent() util.NullJson {
	return util.NullJson{
		Json:  json.RawMessage(_defaultProjectContent),
		Valid: true,
	}
}
