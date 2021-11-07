package dataset

import (
	"github.com/jmoiron/sqlx"
	"nns_back/util"
)

type mysqlRepository struct {
	db *sqlx.DB
}

func NewMysqlRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{
		db: db,
	}
}

//func (m *mysqlRepository) count(builder squirrel.SelectBuilder) (int64, error) {
//	query, args, err := builder.Columns("COUNT(ds.*)").from("dataset ds").ToSql()
//	if err != nil {
//		return 0, err
//	}
//
//	var count int64
//	err = m.db.QueryRowx(query, args...).Scan(&count)
//	return count, err
//}

func (m *mysqlRepository) CountPublic() (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT count(*)
FROM dataset ds
WHERE ds.public = TRUE
  AND ds.status = 'EXIST';
`).Scan(&count)
	return count, err
}

func (m *mysqlRepository) CountPublicByUserName(userName string) (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT count(*)
FROM dataset ds
         JOIN user u on ds.user_id = u.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND u.name = ?;
`, userName).Scan(&count)
	return count, err
}

func (m *mysqlRepository) CountPublicByUserNameLike(userName string) (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT count(*)
FROM dataset ds
         JOIN user u on ds.user_id = u.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND u.name LIKE ?;
`, util.LikeArg(userName)).Scan(&count)
	return count, err
}

func (m *mysqlRepository) CountPublicByTitle(title string) (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT count(*)
FROM dataset ds
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND ds.name = ?;
`, title).Scan(&count)
	return count, err
}

func (m *mysqlRepository) CountPublicByTitleLike(title string) (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT count(*)
FROM dataset ds
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND ds.name LIKE ?;
`, util.LikeArg(title)).Scan(&count)
	return count, err
}

func (m *mysqlRepository) FindAllPublic(userId int64, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN (SELECT idsl.*
                    FROM dataset_library idsl
                    WHERE idsl.user_id = ?) dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
ORDER BY ds.id DESC
LIMIT ?, ?;
`, userId, offset, limit)
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

func (m *mysqlRepository) FindAllPublicByUserName(userId int64, userName string, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN (SELECT idsl.*
                    FROM dataset_library idsl
                    WHERE idsl.user_id = ?) dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
         JOIN user u on ds.user_id = u.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND u.name = ?
ORDER BY ds.id DESC
LIMIT ?, ?;
`, userId, userName, offset, limit)
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

func (m *mysqlRepository) FindAllPublicByUserNameLike(userId int64, userName string, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN (SELECT idsl.*
                    FROM dataset_library idsl
                    WHERE idsl.user_id = ?) dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
         JOIN user u on ds.user_id = u.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND u.name LIKE ?
ORDER BY ds.id DESC
LIMIT ?, ?;
`, userId, util.LikeArg(userName), offset, limit)
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

func (m *mysqlRepository) FindAllPublicByTitle(userId int64, title string, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN (SELECT idsl.*
                    FROM dataset_library idsl
                    WHERE idsl.user_id = ?) dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND ds.name = ?
ORDER BY ds.id DESC
LIMIT ?, ?;
`, userId, title, offset, limit)
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

func (m *mysqlRepository) FindAllPublicByTitleLike(userId int64, title string, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN (SELECT idsl.*
                    FROM dataset_library idsl
                    WHERE idsl.user_id = ?) dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
WHERE ds.public = TRUE
  AND ds.status = 'EXIST'
  AND ds.name LIKE ?
ORDER BY ds.id DESC
LIMIT ?, ?;
`, userId, util.LikeArg(title), offset, limit)
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

func (m *mysqlRepository) FindNextDatasetNo(userId int64) (int64, error) {
	var dsn int64
	err := m.db.QueryRowx(
		`SELECT ds.dataset_no FROM dataset ds WHERE ds.user_id = ? and ds.status != 'DELETED' ORDER BY ds.dataset_no DESC LIMIT 1`, userId).Scan(&dsn)
	return dsn, err
}

func (m *mysqlRepository) FindByID(id int64) (Dataset, error) {
	ds := Dataset{}
	err := m.db.QueryRowx(`
SELECT ds.id,
       ds.user_id,
       ds.dataset_no,
       ds.url,
       ds.origin_url,
       ds.name,
       ds.description,
       ds.public,
       ds.status,
       ds.image_id,
       ds.kind,
       ds.create_time,
       ds.update_time,
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         LEFT JOIN dataset_library dsl on ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
WHERE ds.id = ?
  and ds.status != 'DELETED';
  `, id).StructScan(&ds)

	return ds, err
}

func (m *mysqlRepository) Insert(dataset Dataset) (int64, error) {
	result, err := m.db.NamedExec(
		`INSERT INTO dataset (user_id,
                     dataset_no,
                     url,
                     origin_url,
                     name,
                     description,
                     public,
                     status,
                     image_id,
                     kind,
                     create_time,
                     update_time)
VALUES (:user_id,
        :dataset_no,
        :url,
        :origin_url,
        :name,
        :description,
        :public,
        :status,
        :image_id,
        :kind,
        :create_time,
        :update_time);`, dataset)

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (m *mysqlRepository) Update(id int64, dataset Dataset) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dataset.ID = id
	_, err = tx.NamedExec(`
UPDATE dataset SET user_id = :user_id,
                   dataset_no = :dataset_no,
                   url = :url,
                   origin_url = :origin_url,
                   name = :name,
                   description = :description,
                   public      = :public,
                   status 	   = :status,
                   image_id	   = :image_id,
                   kind        = :kind,
                   create_time = :create_time,
                   update_time = :update_time
WHERE id = :id and status != 'DELETED';
`, dataset)

	usable := dataset.Public.Bool && dataset.Status == EXIST
	err = changeDatasetLibraryUsable(tx, dataset.ID, usable)
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *mysqlRepository) Delete(id int64) error {
	tx, err := m.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE dataset SET status = 'DELETED' WHERE id = ? and status = 'EXIST'`, id)
	if err != nil {
		return err
	}

	err = changeDatasetLibraryUsable(tx, id, false)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func changeDatasetLibraryUsable(tx *sqlx.Tx, datasetId int64, usable bool) error {
	_, err := tx.Exec(`
UPDATE dataset_library dsl
SET dsl.usable = ?
WHERE dsl.dataset_id = ?;
`, usable, datasetId)

	return err
}

func (m *mysqlRepository) FindDatasetFromDatasetLibraryByUserId(userId int64, offset, limit int) ([]Dataset, error) {
	rows, err := m.db.Queryx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         JOIN dataset_library dsl ON ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
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

func (m *mysqlRepository) CountDatasetLibraryByUserId(userId int64) (int64, error) {
	var count int64
	err := m.db.QueryRowx(`
SELECT COUNT(*)
FROM dataset_library dsl
WHERE dsl.user_id = ?;
`, userId).Scan(&count)

	return count, err
}

func (m *mysqlRepository) FindDatasetFromDatasetLibraryByDatasetId(userId int64, datasetId int64) (Dataset, error) {
	var dataset Dataset
	err := m.db.QueryRowx(`
SELECT ds.id              "id",
       ds.user_id         "user_id",
       ds.dataset_no      "dataset_no",
       ds.url             "url",
       ds.origin_url      "origin_url",
       ds.name            "name",
       ds.description     "description",
       ds.public          "public",
       ds.status          "status",
       ds.image_id        "image_id",
       ds.kind            "kind",
       ds.create_time     "create_time",
       ds.update_time     "update_time",
       dsl.usable         "usable",
       dsl.id IS NOT NULL "in_library",
       i.url              "thumbnail_url"
FROM dataset ds
         JOIN dataset_library dsl ON ds.id = dsl.dataset_id
         LEFT JOIN image i on ds.image_id = i.id
WHERE dsl.user_id = ?
  AND dsl.dataset_id = ?;
`, userId, datasetId).StructScan(&dataset)

	return dataset, err
}

func (m *mysqlRepository) AddDatasetToDatasetLibrary(userId int64, datasetId int64) error {
	_, err := m.db.Exec(`
INSERT INTO dataset_library (user_id, dataset_id, usable)
SELECT ? "user_id", ds.id "dataset_id", (ds.public IS TRUE OR ds.user_id = ?) "usable"
FROM dataset ds
WHERE ds.id = ? AND ds.status = 'EXIST';
`, userId, userId, datasetId)

	return err
}

func (m *mysqlRepository) DeleteDatasetFromDatasetLibrary(userId int64, datasetId int64) error {
	_, err := m.db.Exec(`
DELETE dsl
FROM dataset_library dsl
WHERE dsl.user_id = ?
  AND dsl.dataset_id = ?;
`, userId, datasetId)

	return err
}
