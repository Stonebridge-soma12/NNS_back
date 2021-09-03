package dataset

type Repository interface {
	FindByID(id int64) (Dataset, error)
	Insert(dataset Dataset) (int64, error)
	Update(id int64, dataset Dataset) error
	Delete(id int64) error
}

