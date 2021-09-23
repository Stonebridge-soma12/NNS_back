package dataset

import (
	"github.com/jmoiron/sqlx"
)

type MysqlRepository struct {
	DB *sqlx.DB
}

func (m *MysqlRepository) CountPublic() (int64, error) {
	var count int64
	err := m.DB.QueryRowx(`SELECT count(*) FROM dataset ds WHERE ds.public = 1 and ds.status = 'EXIST';`).Scan(&count)
	return count, err
}

func (m *MysqlRepository) CountByUserId(userId int64) (int64, error) {
	var count int64
	err := m.DB.QueryRowx(`SELECT count(*) FROM dataset ds WHERE ds.user_id = ? and ds.status = 'EXIST'`, userId).Scan(&count)
	return count, err
}

func (m *MysqlRepository) FindAllPublic(offset, limit int) ([]Dataset, error) {
	rows, err := m.DB.Queryx(
		`SELECT ds.id,
       ds.user_id,
       ds.dataset_no,
       ds.url,
       ds.name,
       ds.description,
       ds.public,
       ds.status,
       ds.create_time,
       ds.update_time
FROM dataset ds
WHERE ds.public = 1 and ds.status = 'EXIST'
ORDER BY ds.id DESC 
LIMIT ?, ?;`, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dsList := make([]Dataset, 0)
	for rows.Next() {
		ds := Dataset{}
		if err := rows.StructScan(&ds); err != nil {
			return nil, err
		}

		dsList = append(dsList, ds)
	}

	return dsList, nil
}

func (m *MysqlRepository) FindNextDatasetNo(userId int64) (int64, error) {
	var dsn int64
	err := m.DB.QueryRowx(
		`SELECT ds.dataset_no FROM dataset ds WHERE ds.user_id = ? and ds.status != 'DELETED' ORDER BY ds.dataset_no DESC LIMIT 1`, userId).Scan(&dsn)
	return dsn, err
}

func (m *MysqlRepository) FindByUserId(userId int64, offset, limit int) ([]Dataset, error) {
	rows, err := m.DB.Queryx(
		`SELECT ds.id,
       ds.user_id,
       ds.dataset_no,
       ds.url,
       ds.name,
       ds.description,
       ds.public,
       ds.status,
       ds.create_time,
       ds.update_time
FROM dataset ds
WHERE ds.user_id = ? and ds.status = 'EXIST'
ORDER BY ds.id DESC 
LIMIT ?, ?;`, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dsList := make([]Dataset, 0)
	for rows.Next() {
		ds := Dataset{}
		if err := rows.StructScan(&ds); err != nil {
			return nil, err
		}

		dsList = append(dsList, ds)
	}

	return dsList, nil
}

func (m *MysqlRepository) FindByID(id int64) (Dataset, error) {
	ds := Dataset{}
	err := m.DB.QueryRowx(
		`SELECT ds.id,
       ds.user_id,
       ds.dataset_no,
       ds.url,
       ds.name,
       ds.description,
       ds.public,
       ds.status,
       ds.create_time,
       ds.update_time
FROM dataset ds
WHERE ds.id = ?
  and ds.status != 'DELETED';`, id).StructScan(&ds)

	return ds, err
}

func (m *MysqlRepository) FindByUserIdAndDatasetNo(userId int64, datasetNo int64) (Dataset, error) {
	ds := Dataset{}
	err := m.DB.QueryRowx(
		`SELECT ds.id,
       ds.user_id,
       ds.dataset_no,
       ds.url,
       ds.name,
       ds.description,
       ds.public,
       ds.status,
       ds.create_time,
       ds.update_time
FROM dataset ds
WHERE ds.user_id = ? AND ds.dataset_no = ?
  AND ds.status != 'DELETED';`, userId, datasetNo).StructScan(&ds)

	return ds, err
}

func (m *MysqlRepository) Insert(dataset Dataset) (int64, error) {
	result, err := m.DB.NamedExec(
		`INSERT INTO dataset (user_id,
                     dataset_no,
                     url,
                     name,
                     description,
                     public,
                     status,
                     create_time,
                     update_time)
VALUES (:user_id,
        :dataset_no,
        :url,
        :name,
        :description,
        :public,
        :status,
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
                   dataset_no = :dataset_no,
                   url = :url,
                   name = :name,
                   description = :description,
                   public      = :public,
                   status 	   = :status,
                   create_time = :create_time,
                   update_time = :update_time
WHERE id = :id and status != 'DELETED';`, dataset)

	return err
}

func (m *MysqlRepository) Delete(id int64) error {
	_, err := m.DB.Exec(`UPDATE dataset SET status = 'DELETED' WHERE id = ? and status = 'EXIST'`, id)
	return err
}

func (m *MysqlRepository) FindDatasetFromDatasetLibraryByUserId(userId int64, offset, limit int) ([]Dataset, error) {
	rows, err := m.DB.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library"
FROM dataset ds
         JOIN dataset_library dsl ON ds.id = dsl.dataset_id
WHERE dsl.user_id = ?
ORDER BY dsl.create_time DESC
LIMIT ?, ?;
`, userId, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	datasets := make([]Dataset, 0)
	for rows.Next() {
		dataset := Dataset{}
		if err := rows.StructScan(&dataset); err != nil {
			return nil, err
		}

		datasets = append(datasets, dataset)
	}

	return datasets, nil
}

func (m *MysqlRepository) CountDatasetLibraryByUserId(userId int64) (int64, error) {
	var count int64
	err := m.DB.QueryRowx(`
SELECT COUNT(*)
FROM dataset_library dsl
WHERE dsl.user_id = ?;
`, userId).Scan(&count)

	return count, err
}

func (m *MysqlRepository) FindDatasetFromDatasetLibraryByDatasetId(userId int64, datasetId int64) (Dataset, error) {
	var dataset Dataset
	err := m.DB.QueryRowx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library"
FROM dataset ds
         JOIN dataset_library dsl ON ds.id = dsl.dataset_id
WHERE dsl.user_id = ?
  AND dsl.dataset_id = ?;
`, userId, datasetId).StructScan(&dataset)

	return dataset, err
}

func (m *MysqlRepository) AddDatasetToDatasetLibrary(userId int64, datasetId int64) error {
	_, err := m.DB.Exec(`
INSERT INTO dataset_library (user_id, dataset_id, usable)
SELECT ? "user_id", ds.id "dataset_id", ds.status = 'EXIST' AND ds.public IS TRUE "usable"
FROM dataset ds
WHERE ds.id = ?;
`, userId, datasetId)

	return err
}

func (m *MysqlRepository) DeleteDatasetFromDatasetLibrary(userId int64, datasetId int64) error {
	_, err := m.DB.Exec(`
DELETE dsl
FROM dataset_library dsl
WHERE dsl.user_id = ?
  AND dsl.dataset_id = ?;
`, userId, datasetId)

	return err
}