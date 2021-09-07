package dataset

import "github.com/jmoiron/sqlx"

type MysqlRepository struct {
	DB *sqlx.DB
}

func (m *MysqlRepository) FindByID(id int64) (Dataset, error) {
	ds := Dataset{}
	err := m.DB.QueryRowx(
		`SELECT ds.id, 
       ds.user_id, 
       ds.url, 
       ds.name, 
       ds.description, 
       ds.create_time, 
       ds.update_time
FROM dataset ds
WHERE ds.id = ?;`, id).StructScan(&ds)

	return ds, err
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
	dataset.ID = id
	_, err := m.DB.NamedExec(
		`UPDATE dataset SET user_id = :user_id,
                   url = :url,
                   name = :name,
                   description = :description,
                   create_time = :create_time,
                   update_time = :update_time
WHERE id = :id;`,dataset)

	return err
}

func (m *MysqlRepository) Delete(id int64) error {
	// TODO: implement
	return nil
}