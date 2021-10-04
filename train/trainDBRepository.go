package train

import (
	"github.com/jmoiron/sqlx"
)

const (
	defaultSelectTrainHistoryColumns = `t.id,
								   t.user_id,
								   t.train_no,
								   t.project_id,
								   t.acc,
								   t.loss,
								   t.val_acc,
								   t.val_loss,
								   t.name,
								   t.epochs,
								   t.result_url,
								   t.status,
								   tc.id,
								   tc.train_id,
								   tc.train_dataset_url,
								   tc.valid_dataset_url,
								   tc.dataset_shuffle,
								   tc.dataset_label,
								   tc.dataset_normalization_usage,
								   tc.dataset_normalization_method,
								   tc.model_content,
								   tc.model_config,
								   tc.create_time,
								   tc.update_time`
	fromTrain = `train t`
	fromTrainConfig = `train_config tc`
	defaultDeleteTrainQuery = "update train set status='DEL' "
	defaultSelectTrainQuery = `SELECT t.id,
								   t.user_id,
								   t.train_no,
								   t.project_id,
								   t.acc,
								   t.loss,
								   t.val_acc,
								   t.val_loss,
								   t.name,
								   t.epochs,
								   t.url,
								   t.status,
								   tc.id,
								   tc.train_id,
								   tc.train_dataset_url,
								   tc.valid_dataset_url,
								   tc.dataset_shuffle,
								   tc.dataset_label,
								   tc.dataset_normalization_usage,
								   tc.dataset_normalization_method,
								   tc.model_content,
								   tc.model_config,
								   tc.create_time,
								   tc.update_time
							FROM train t
									 JOIN train_config tc ON t.id = tc.train_id
									 JOIN project p ON t.project_id = p.id
							WHERE p.user_id = ?
							  AND p.project_no = ?
							  AND t.status != 'DEL'
							LIMIT ?, ?`
)

type TrainDbRepository struct {
	DB *sqlx.DB
}

func WithoutTrainStatus(status string) Option {
	return optionFunc(func(o *options) {
		o.queryString += `t.status != ? `
		o.args = append(o.args, status)
	})
}

func (tdb *TrainDbRepository) FindNextTrainNo(userId int64) (int64, error) {
	var lastTrainNo int64
	err := tdb.DB.QueryRowx(`
SELECT t.train_no
FROM train t
WHERE t.user_id = ?
ORDER BY t.train_no DESC
LIMIT 1;
`, userId).Scan(&lastTrainNo)

	return lastTrainNo + 1, err
}

func (tdb *TrainDbRepository) CountCurrentTraining(userId int64) (int, error) {
	var count int
	err := tdb.DB.QueryRowx(`
SELECT COUNT(*)
FROM train t
WHERE t.user_id = ?
  AND t.status = 'TRAIN';
`, userId).Scan(&count)

	return count, err
}

func insertTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into " +
			"train (status, acc, loss, val_acc, val_loss, epochs, name, result_url) " +
			"values(:status, :acc, :loss, :val_acc, :val_loss, :epochs, :name, :result_url)"
	})
}

func updateTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "update train " +
			"set status=:status, acc=:acc, loss=:loss, val_acc=:val_acc, val_loss=:val_loss, epochs=:epochs, name=:name, result_url=:result_url " +
			"where id = :id"
	})
}

func WithUserId(userId int64) Option {
	return optionFunc(func(o *options) {
		o.queryString += `WHERE user_id = ? `
		o.args = append(o.args, userId)
	})
}

func (tdb *TrainDbRepository) Insert(train Train) (int64, error) {
	//options := options{}
	//ApplyOptions(&options, insertTrain())
	//
	//result, err := tdb.DB.NamedExec(options.queryString, &train)
	//if err != nil {
	//	return 0, err
	//}

	tx, err := tdb.DB.Beginx()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	result, err := tx.NamedExec(`
INSERT INTO train (user_id,
                   train_no,
                   project_id,
                   acc,
                   loss,
                   val_acc,
                   val_loss,
                   name,
                   epochs,
                   result_url,
                   status)
VALUES (:user_id,
        :train_no,
        :project_id,
        :acc,
        :loss,
        :val_acc,
        :val_loss,
        :name,
        :epochs,
        :result_url,
        :status);
`, train)
	if err != nil {
		return 0, err
	}

	insertedTrainId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	train.TrainConfig.TrainId = insertedTrainId

	_, err = tx.NamedExec(`
INSERT INTO train_config (train_id,
                          train_dataset_url,
                          valid_dataset_url,
                          dataset_shuffle,
                          dataset_label,
                          dataset_normalization_usage,
                          dataset_normalization_method,
                          model_content,
                          model_config)
VALUES (:train_id,
        :train_dataset_url,
        :valid_dataset_url,
        :dataset_shuffle,
        :dataset_label,
        :dataset_normalization_usage,
        :dataset_normalization_method,
        :model_content,
        :model_config);
`, train.TrainConfig)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return insertedTrainId, nil
}

func (tdb *TrainDbRepository) Delete(opts ...Option) error {
	options := options{
		queryString: defaultDeleteTrainQuery,
	}
	ApplyOptions(&options, opts...)

	_, err := tdb.DB.Exec(options.queryString, options.args)
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TrainDbRepository) Update(train Train, opts ...Option) error {
	options := options{}
	ApplyOptions(&options, updateTrain())

	_, err := tdb.DB.NamedExec(options.queryString, &train)
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TrainDbRepository) Find(opts ...Option) (Train, error) {
	options := options{
		queryString: defaultSelectTrainQuery,
	}
	ApplyOptions(&options, opts...)

	var train Train
	err := tdb.DB.Get(&train, options.queryString, options.args...)
	if err != nil {
		return Train{}, err
	}

	return train, err
}

func (tdb *TrainDbRepository) FindAll(opts ...Option) ([]Train, error) {
	options := options{
		queryString: defaultSelectTrainQuery,
	}
	ApplyOptions(&options, opts...)

	var trains []Train
	rows, err := tdb.DB.Queryx(options.queryString, options.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var train Train
		err := rows.StructScan(&train)
		if err != nil {
			return nil, err
		}
		trains = append(trains, train)
	}

	return trains, nil
}

func (tdb *TrainDbRepository) GetTrainingCount(userId int64) (int, error) {
	var count int
	err := tdb.DB.QueryRowx(`
SELECT COUNT(*)
FROM train t
WHERE t.user_id = ? AND t.status = 'TRAIN';
`, userId).Scan(&count)

	return count, err
}
