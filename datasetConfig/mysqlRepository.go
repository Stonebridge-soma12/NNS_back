package datasetConfig

import (
	"github.com/jmoiron/sqlx"
	"nns_back/util"
)

type mysqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) FindByProjectIdAndDatasetConfigName(projectId int64, datasetConfigName string) (DatasetConfig, error) {
	var result DatasetConfig
	err := r.db.QueryRowx(`
SELECT dc.id,
       dc.project_id,
       dc.dataset_id,
       dc.name,
       dc.shuffle,
       dc.label,
       dc.normalization_method,
       dc.status,
       dc.create_time,
       dc.update_time
FROM dataset_config dc
         JOIN project p on dc.project_id = p.id
WHERE p.id = ?
  AND dc.name = ?
  AND dc.status = 'EXIST';`, projectId, datasetConfigName).StructScan(&result)
	return result, err
}

func (r *mysqlRepository) CountByProjectId(projectId int64) (int64, error) {
	var count int64
	err := r.db.QueryRowx(`
SELECT COUNT(dc.id)
FROM dataset_config dc
         JOIN project p on dc.project_id = p.id
WHERE p.id = ? AND dc.status = 'EXIST';`, projectId).Scan(&count)

	return count, err
}

func (r *mysqlRepository) FindAllByProjectId(projectId int64, offset int, limit int) ([]DatasetConfig, error) {
	rows, err := r.db.Queryx(`
SELECT dc.id,
       dc.project_id,
       dc.dataset_id,
       dc.name,
       dc.shuffle,
       dc.label,
       dc.normalization_method,
       dc.status,
       dc.create_time,
       dc.update_time,
       d.name "dataset_name"
FROM dataset_config dc
         JOIN project p on dc.project_id = p.id
         JOIN dataset d on dc.dataset_id = d.id
WHERE p.id = ?
  AND dc.status = 'EXIST'
ORDER BY dc.id ASC
LIMIT ?, ?;`, projectId, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	datasetConfigList := make([]DatasetConfig, 0)
	for rows.Next() {
		var dc DatasetConfig
		if err := rows.StructScan(&dc); err != nil {
			return nil, err
		}

		datasetConfigList = append(datasetConfigList, dc)
	}

	return datasetConfigList, nil
}

func (r *mysqlRepository) FindByUserIdAndId(userId int64, id int64) (DatasetConfig, error) {
	var result DatasetConfig
	err := r.db.QueryRowx(`
SELECT dc.id,
       dc.project_id,
       dc.dataset_id,
       dc.name,
       dc.shuffle,
       dc.label,
       dc.normalization_method,
       dc.status,
       dc.create_time,
       dc.update_time,
       d.name "dataset_name"
FROM dataset_config dc
         JOIN project p on dc.project_id = p.id
         JOIN user u on p.user_id = u.id
         JOIN dataset d on dc.dataset_id = d.id
WHERE u.id = ?
  AND dc.id = ?
  AND dc.status = 'EXIST';`, userId, id).StructScan(&result)

	return result, err
}

func (r *mysqlRepository) Insert(datasetConfig DatasetConfig) (int64, error) {
	result, err := r.db.NamedExec(`
INSERT INTO dataset_config (project_id,
                            dataset_id,
                            name,
                            shuffle,
                            label,
                            normalization_method,
                            status)
VALUES (:project_id,
        :dataset_id,
        :name,
        :shuffle,
        :label,
        :normalization_method,
        :status);`, datasetConfig)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *mysqlRepository) Update(datasetConfig DatasetConfig) error {
	_, err := r.db.NamedExec(`
UPDATE dataset_config
SET project_id           = :project_id,
    dataset_id           = :dataset_id,
    name                 = :name,
    shuffle              = :shuffle,
    label                = :label,
    normalization_method = :normalization_method,
    status               = :status
WHERE id = :id;`, datasetConfig)
	return err
}

func (r *mysqlRepository) Delete(datasetConfig DatasetConfig) error {
	datasetConfig.Status = util.StatusDELETED
	return r.Update(datasetConfig)
}
