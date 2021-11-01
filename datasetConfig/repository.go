package datasetConfig

type Repository interface {
	CountByProjectId(projectId int64) (int64, error)
	FindAllByProjectId(projectId int64, offset int, limit int) ([]DatasetConfig, error)
	FindByUserIdAndId(userId int64, id int64) (DatasetConfig, error)
	FindByProjectIdAndDatasetConfigName(userId int64, datasetConfigName string) (DatasetConfig, error)
	Insert(datasetConfig DatasetConfig) (int64, error)
	Update(datasetConfig DatasetConfig) error
	Delete(datasetConfig DatasetConfig) error
}
