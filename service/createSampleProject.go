package service

import (
	"database/sql"
	"encoding/json"
	"nns_back/dataset"
	"nns_back/datasetConfig"
	"nns_back/model"
	"nns_back/repository"
	"nns_back/util"
	"time"
)

const (
	_sampleProjectNo          = 1
	_sampleProjectName        = "Sample project"
	_sampleProjectDescription = "자동으로 생성되는 샘플 프로젝트입니다. MNIST 데이터셋을 사용한 손글씨 분류 모델이 구성되어 있습니다."
	_sampleProjectConfig      = `{
    "batch_size": 32,
    "dataset_config": {
        "id": 0,
        "valid": false
    },
    "early_stop": {
        "monitor": "loss",
        "patience": 2,
        "usage": true
    },
    "epochs": 10,
    "learning_rate_reduction": {
        "factor": 0.25,
        "min_lr": 3e-7,
        "monitor": "val_accuracy",
        "patience": 2,
        "usage": true
    },
    "loss": "categorical_crossentropy",
    "metrics": [
        "accuracy"
    ],
    "optimizer_config": {
        "amsgrad": false,
        "beta_1": 0.9,
        "beta_2": 0.999,
        "centered": false,
        "decay": 1,
        "epsilon": 1e-7,
        "initial_accumulator_value": 1,
        "learning_rate": 0.001,
        "momentum": 1,
        "nesterov": false,
        "weight_decay": 1
    },
    "optimizer_name": "Adam"
}`
	_sampleProjectContent = `{
    "flowState": {
        "elements": [
            {
                "data": {
                    "category": "Layer",
                    "label": "InputNode_1",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            28,
                            28,
                            1
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Input"
                },
                "id": "node_default_input_node_auto_created",
                "position": {
                    "x": 168.140625,
                    "y": -32
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Conv2D_m0",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 32,
                        "kernel_size": [
                            3,
                            3
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            1,
                            1
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Conv2D"
                },
                "id": "node_643c9e102a74420184c7b5331c4ebe358",
                "position": {
                    "x": 338.4195667547598,
                    "y": 117.94425593010078
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Activation_43",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Activation"
                },
                "id": "node_321d67798be14143ae8b7c8dbbc6c10b8",
                "position": {
                    "x": 126.62033090987549,
                    "y": 231.460950649911
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "MaxPool2D_Uf",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            2,
                            2
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            1,
                            1
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "MaxPool2D"
                },
                "id": "node_c4be4e574c274aa294290ba3bfe566708",
                "position": {
                    "x": 124.49766545493775,
                    "y": 334.3159305646946
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Conv2D_bA",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 64,
                        "kernel_size": [
                            3,
                            3
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            1,
                            1
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Conv2D"
                },
                "id": "node_5fd4f3eec42e44bdb7e1dfad473286798",
                "position": {
                    "x": 137.640625,
                    "y": 440
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Activation_cN",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Activation"
                },
                "id": "node_07a3ba6919584fc484989b86309313c38",
                "position": {
                    "x": 233.64062499999997,
                    "y": 538
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Flatten_hD",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Flatten"
                },
                "id": "node_86d4bbf7730641d1988bc29f42a4fb298",
                "position": {
                    "x": 236.64062499999997,
                    "y": 689
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Activation_VS",
                    "param": {
                        "activation": "softmax",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 0
                    },
                    "type": "Activation"
                },
                "id": "node_c4d2a7367a4a4a3982c3f81e6ec13c138",
                "position": {
                    "x": 335.640625,
                    "y": 966.9999999999999
                },
                "type": "Layer"
            },
            {
                "data": {
                    "category": "Layer",
                    "label": "Dense_hr",
                    "param": {
                        "activation": "relu",
                        "axis": 0,
                        "comment": "",
                        "epsilon": 0,
                        "filters": 0,
                        "kernel_size": [
                            0,
                            0
                        ],
                        "momentum": 0,
                        "offset": 0,
                        "padding": "Same",
                        "pool_size": [
                            0,
                            0
                        ],
                        "rate": 0.1,
                        "scale": 0,
                        "shape": [
                            0,
                            0
                        ],
                        "strides": [
                            0,
                            0
                        ],
                        "target_shape": 0,
                        "units": 10
                    },
                    "type": "Dense"
                },
                "id": "node_cbe3cfd184bd487a8dc3fe6498e3104c15",
                "position": {
                    "x": 401.9780864870031,
                    "y": 827.3859576617187
                },
                "type": "Layer"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_default_input_node_auto_creatednull-node_643c9e102a74420184c7b5331c4ebe358null",
                "source": "node_default_input_node_auto_created",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_643c9e102a74420184c7b5331c4ebe358",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_643c9e102a74420184c7b5331c4ebe358null-node_321d67798be14143ae8b7c8dbbc6c10b8null",
                "source": "node_643c9e102a74420184c7b5331c4ebe358",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_321d67798be14143ae8b7c8dbbc6c10b8",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_321d67798be14143ae8b7c8dbbc6c10b8null-node_c4be4e574c274aa294290ba3bfe566708null",
                "source": "node_321d67798be14143ae8b7c8dbbc6c10b8",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_c4be4e574c274aa294290ba3bfe566708",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_c4be4e574c274aa294290ba3bfe566708null-node_5fd4f3eec42e44bdb7e1dfad473286798null",
                "source": "node_c4be4e574c274aa294290ba3bfe566708",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_5fd4f3eec42e44bdb7e1dfad473286798",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_5fd4f3eec42e44bdb7e1dfad473286798null-node_07a3ba6919584fc484989b86309313c38null",
                "source": "node_5fd4f3eec42e44bdb7e1dfad473286798",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_07a3ba6919584fc484989b86309313c38",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_07a3ba6919584fc484989b86309313c38null-node_86d4bbf7730641d1988bc29f42a4fb298null",
                "source": "node_07a3ba6919584fc484989b86309313c38",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_86d4bbf7730641d1988bc29f42a4fb298",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_86d4bbf7730641d1988bc29f42a4fb298null-node_cbe3cfd184bd487a8dc3fe6498e3104c15null",
                "source": "node_86d4bbf7730641d1988bc29f42a4fb298",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_cbe3cfd184bd487a8dc3fe6498e3104c15",
                "targetHandle": null,
                "type": "default"
            },
            {
                "animated": true,
                "id": "reactflow__edge-node_cbe3cfd184bd487a8dc3fe6498e3104c15null-node_c4d2a7367a4a4a3982c3f81e6ec13c138null",
                "source": "node_cbe3cfd184bd487a8dc3fe6498e3104c15",
                "sourceHandle": null,
                "style": {
                    "cursor": "pointer",
                    "stroke": "black",
                    "strokeWidth": 4
                },
                "target": "node_c4d2a7367a4a4a3982c3f81e6ec13c138",
                "targetHandle": null,
                "type": "default"
            }
        ],
        "position": [
            197.9417921516109,
            111.52368720168295
        ],
        "zoom": 0.870550563296124
    },
    "input": "inputnode_1",
    "layers": [
        {
            "category": "Layer",
            "id": "node_default_input_node_auto_created",
            "input": [],
            "name": "inputnode_1",
            "output": [
                "conv2d_m0"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    28,
                    28,
                    1
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Input"
        },
        {
            "category": "Layer",
            "id": "node_643c9e102a74420184c7b5331c4ebe358",
            "input": [
                "inputnode_1"
            ],
            "name": "conv2d_m0",
            "output": [
                "activation_43"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 32,
                "kernel_size": [
                    3,
                    3
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    1,
                    1
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Conv2D"
        },
        {
            "category": "Layer",
            "id": "node_321d67798be14143ae8b7c8dbbc6c10b8",
            "input": [
                "conv2d_m0"
            ],
            "name": "activation_43",
            "output": [
                "maxpool2d_uf"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Activation"
        },
        {
            "category": "Layer",
            "id": "node_c4be4e574c274aa294290ba3bfe566708",
            "input": [
                "activation_43"
            ],
            "name": "maxpool2d_uf",
            "output": [
                "conv2d_ba"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    2,
                    2
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    1,
                    1
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "MaxPool2D"
        },
        {
            "category": "Layer",
            "id": "node_5fd4f3eec42e44bdb7e1dfad473286798",
            "input": [
                "maxpool2d_uf"
            ],
            "name": "conv2d_ba",
            "output": [
                "activation_cn"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 64,
                "kernel_size": [
                    3,
                    3
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    1,
                    1
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Conv2D"
        },
        {
            "category": "Layer",
            "id": "node_07a3ba6919584fc484989b86309313c38",
            "input": [
                "conv2d_ba"
            ],
            "name": "activation_cn",
            "output": [
                "flatten_hd"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Activation"
        },
        {
            "category": "Layer",
            "id": "node_86d4bbf7730641d1988bc29f42a4fb298",
            "input": [
                "activation_cn"
            ],
            "name": "flatten_hd",
            "output": [
                "dense_hr"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Flatten"
        },
        {
            "category": "Layer",
            "id": "node_c4d2a7367a4a4a3982c3f81e6ec13c138",
            "input": [
                "dense_hr"
            ],
            "name": "activation_vs",
            "output": [],
            "param": {
                "activation": "softmax",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 0
            },
            "type": "Activation"
        },
        {
            "category": "Layer",
            "id": "node_cbe3cfd184bd487a8dc3fe6498e3104c15",
            "input": [
                "flatten_hd"
            ],
            "name": "dense_hr",
            "output": [
                "activation_vs"
            ],
            "param": {
                "activation": "relu",
                "axis": 0,
                "comment": "",
                "epsilon": 0,
                "filters": 0,
                "kernel_size": [
                    0,
                    0
                ],
                "momentum": 0,
                "offset": 0,
                "padding": "Same",
                "pool_size": [
                    0,
                    0
                ],
                "rate": 0.1,
                "scale": 0,
                "shape": [
                    0,
                    0
                ],
                "strides": [
                    0,
                    0
                ],
                "target_shape": 0,
                "units": 10
            },
            "type": "Dense"
        }
    ],
    "output": "activation_vs"
}`
)

const (
	_sampleDatasetId = 28
)

const (
	_sampleDatasetConfigName = "Use MNIST dataset configuration"
	_sampleDatasetConfigShuffle = true
	_sampleDatasetConfigNormalizationMethod = "IMAGE"
	_sampleDatasetConfigLabel = "label"
)

func CreateAutoCreatedSampleProject(userId int64, projectRepo repository.ProjectRepository, datasetRepo dataset.Repository, datasetConfigRepo datasetConfig.Repository) error {
	if err := addSampleDatasetLibrary(userId, datasetRepo); err != nil {
		return err
	}

	sampleProjectId, err := addSampleProject(userId, projectRepo)
	if err != nil {
		return err
	}

	datasetConfigId, err := addSampleDatasetConfig(sampleProjectId, datasetConfigRepo)
	if err != nil {
		return err
	}

	err = updateDatasetConfigValueOfSampleProjectConfig(userId, datasetConfigId, projectRepo)
	if err != nil {
		return err
	}

	return nil
}

func addSampleDatasetLibrary(userId int64, repo dataset.Repository) error {
	return repo.AddDatasetToDatasetLibrary(userId, _sampleDatasetId)
}

func addSampleProject(userId int64, repo repository.ProjectRepository) (int64, error) {
	sampleProject := model.Project{
		ShareKey:    sql.NullString{},
		UserId:      userId,
		ProjectNo:   _sampleProjectNo,
		Name:        _sampleProjectName,
		Description: _sampleProjectDescription,
		Config: util.NullJson{
			Json:  json.RawMessage(_sampleProjectConfig),
			Valid: true,
		},
		Content: util.NullJson{
			Json:  json.RawMessage(_sampleProjectContent),
			Valid: true,
		},
		Status:     util.StatusEXIST,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	return repo.Insert(sampleProject)
}

func addSampleDatasetConfig(projectId int64, repo datasetConfig.Repository) (int64, error) {
	sampleDatasetConfig := datasetConfig.DatasetConfig{
		ProjectId:           projectId,
		DatasetId:           _sampleDatasetId,
		Name:                _sampleDatasetConfigName,
		Shuffle:             _sampleDatasetConfigShuffle,
		NormalizationMethod: sql.NullString{
			Valid: true,
			String: _sampleDatasetConfigNormalizationMethod,
		},
		Label:               _sampleDatasetConfigLabel,
		Status:              util.StatusEXIST,
		CreateTime:          time.Now(),
		UpdateTime:          time.Now(),
	}

	return repo.Insert(sampleDatasetConfig)
}

func updateDatasetConfigValueOfSampleProjectConfig(userId, datasetConfigId int64, repo repository.ProjectRepository) error {
	sampleProject, err := repo.SelectProject(repository.ClassifiedByProjectNo(userId, _sampleProjectNo))
	if err != nil {
		return err
	}

	config := make(map[string]interface{})
	if err := json.Unmarshal(sampleProject.Config.Json, &config); err != nil {
		return err
	}

	config["dataset_config"].(map[string]interface{})["id"] = datasetConfigId
	config["dataset_config"].(map[string]interface{})["valid"] = true

	sampleProject.Config.Json, err = json.Marshal(config)
	if err != nil {
		return err
	}

	return repo.Update(sampleProject)
}