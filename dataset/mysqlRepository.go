package dataset

import "github.com/jmoiron/sqlx"

type MysqlRepository struct {
	DB *sqlx.DB
}

func (m *MysqlRepository) FindByID(id int64) (Dataset, error) {
	return Dataset{}, nil
}

func (m *MysqlRepository) Insert(dataset Dataset) (int64, error) {
	result, err := m.DB.NamedExec(
		`INSERT INTO dataset (user_id,
                     url,
                     name,
                     description,
                     create_time,
                     update_time)
VALUES (:user_id,
        :url,
        :name,
        :description,
        :create_time,
        :update_time);`, dataset)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (m *MysqlRepository) Update(id int64, dataset Dataset) error {
	return nil
}

func (m *MysqlRepository) Delete(id int64) error {
	return nil
}