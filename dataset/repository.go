package dataset

type Repository interface {
	FindNextDatasetNo(userId int64) (int64, error)
	FindByID(id int64) (Dataset, error)
	FindByUserIdAndDatasetNo(userId int64, datasetNo int64) (Dataset, error)
	Insert(dataset Dataset) (int64, error)
	Update(id int64, dataset Dataset) error
	Delete(id int64) error

	// dataset list
	CountPublic() (int64, error)
	CountPublicByUserName(userName string) (int64, error)
	CountPublicByUserNameLike(userName string) (int64, error)
	CountPublicByTitle(title string) (int64, error)
	CountPublicByTitleLike(title string) (int64, error)
	FindAllPublic(userId int64, offset, limit int) ([]Dataset, error)
	FindAllPublicByUserName(userId int64, userName string, offset, limit int) ([]Dataset, error)
	FindAllPublicByUserNameLike(userId int64, userName string, offset, limit int) ([]Dataset, error)
	FindAllPublicByTitle(userId int64, title string, offset, limit int) ([]Dataset, error)
	FindAllPublicByTitleLike(userId int64, title string, offset, limit int) ([]Dataset, error)

	// dataset library features
	FindDatasetFromDatasetLibraryByUserId(userId int64, offset, limit int) ([]Dataset, error)
	CountDatasetLibraryByUserId(userId int64) (int64, error)
	FindDatasetFromDatasetLibraryByDatasetId(userId int64, datasetId int64) (Dataset, error)
	AddDatasetToDatasetLibrary(userId int64, datasetId int64) error
	DeleteDatasetFromDatasetLibrary(userId int64, datasetId int64) error
}
