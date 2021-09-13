package dataset

type Repository interface {
	CountPublic() (int64, error)
	CountByUserId(userId int64) (int64, error)
	FindAllPublic(offset, limit int) ([]Dataset, error)
	FindByUserId(userId int64, offset, limit int) ([]Dataset, error)
	FindByID(id int64) (Dataset, error)
	Insert(dataset Dataset) (int64, error)
	Update(id int64, dataset Dataset) error
	Delete(id int64) error
}
