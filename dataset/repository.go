package dataset

type Repository interface {
	FindNextDatasetNo(userId int64) (int64, error)
	CountPublic() (int64, error)
	CountByUserId(userId int64) (int64, error)
	FindAllPublic(offset, limit int) ([]Dataset, error)
	FindByUserId(userId int64, offset, limit int) ([]Dataset, error)
	FindByID(id int64) (Dataset, error)
	FindByUserIdAndDatasetNo(userId int64, datasetNo int64) (Dataset, error)
	Insert(dataset Dataset) (int64, error)
	Update(id int64, dataset Dataset) error
	Delete(id int64) error

	FindDatasetFromDatasetLibraryByUserId(userId int64, offset, limit int) ([]Dataset, error)
	CountDatasetLibraryByUserId(userId int64) (int64, error)
	FindDatasetFromDatasetLibraryByDatasetId(userId int64, datasetId int64) (Dataset, error)
	AddDatasetToDatasetLibrary(userId int64, datasetId int64) error
	DeleteDatasetFromDatasetLibrary(userId int64, datasetId int64) error
}