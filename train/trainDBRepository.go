package train

import (
	"github.com/jmoiron/sqlx"
	"nns_back/query"
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

func WithTrainId(trainId int64) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("train.train_id = ?", trainId)
	})
}

func WithProjectUserId(userId int64) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("p.user_id = ?", userId)
	})
}

func WithProjectProjectNo(projectNo int) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("p.project_no = ?", projectNo)
	})
}

func WithoutTrainStatusDel() query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("t.status != 'DEL'")
	})
}

func WithPagenation(offset, limit int) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddLimit(offset, limit)
	})
}

func WithTrainTrainNo(trainNo int) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("t.train_no = ?", trainNo)
	})
}

func WithTrainUserId(userId int64) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("t.user_id = ?", userId)
	})
}

type TrainDbRepository struct {
	DB *sqlx.DB
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

func (tdb *TrainDbRepository) Delete(opts ...query.Option) error {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddUpdate("train t", "t.status = ?", TrainStatusDelete).
		AddJoin("project p on train.project_id = project.id")

	err := builder.Build()
	if err != nil {
		return err
	}

	_, err = tdb.DB.Exec(builder.QueryString, builder.Args)
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TrainDbRepository) Update(train Train, opts ...query.Option) error {
	builder := query.ApplyQueryOptions()
	builder.AddUpdate(
		"train",
		"status = ?, acc = ?, loss = ?, val_acc = ?, val_loss = ?, epochs = ?, name = ?, result_url = ?",
		train.Status,
		train.Acc,
		train.Loss,
		train.ValAcc,
		train.ValLoss,
		train.Epochs,
		train.Name,
		train.ResultUrl,
	).AddWhere("id = ?", train.Id)

	err := builder.Build()
	if err != nil {
		return nil
	}

	_, err = tdb.DB.Exec(builder.QueryString, builder.Args...)
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TrainDbRepository) Find(opts ...query.Option) (Train, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectTrainHistoryColumns).
		AddFrom("train t").
		AddJoin("train_config tc ON t.id = tc.train_id").
		AddJoin("project p ON t.project_id = p.id")

	builder.Build()

	var train Train
	row := tdb.DB.QueryRow(builder.QueryString, builder.Args...)
	err := row.Scan(
		&train.Id,
		&train.UserId,
		&train.TrainNo,
		&train.ProjectId,
		&train.Acc,
		&train.Loss,
		&train.ValAcc,
		&train.ValLoss,
		&train.Name,
		&train.Epochs,
		&train.ResultUrl,
		&train.Status,
		&train.TrainConfig.Id,
		&train.TrainConfig.TrainId,
		&train.TrainConfig.TrainDatasetUrl,
		&train.TrainConfig.ValidDatasetUrl,
		&train.TrainConfig.DatasetShuffle,
		&train.TrainConfig.DatasetLabel,
		&train.TrainConfig.DatasetNormalizationUsage,
		&train.TrainConfig.DatasetNormalizationMethod,
		&train.TrainConfig.ModelContent,
		&train.TrainConfig.ModelConfig,
		&train.TrainConfig.CreateTime,
		&train.TrainConfig.UpdateTime,
	)

	if err != nil {
		return train, err
	}

	return train, nil
}

func (tdb *TrainDbRepository) FindAll(opts ...query.Option) ([]Train, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectTrainHistoryColumns).
		AddFrom("train t").
		AddJoin("train_config tc ON t.id = tc.train_id").
		AddJoin("project p ON t.project_id = p.id")

	err := builder.Build()
	if err != nil {
		return nil, err
	}

	rows, err := tdb.DB.Queryx(builder.QueryString, builder.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainList []Train
	for rows.Next() {
		var train Train
		err = rows.Scan(
			&train.Id,
			&train.UserId,
			&train.TrainNo,
			&train.ProjectId,
			&train.Acc,
			&train.Loss,
			&train.ValAcc,
			&train.ValLoss,
			&train.Name,
			&train.Epochs,
			&train.ResultUrl,
			&train.Status,
			&train.TrainConfig.Id,
			&train.TrainConfig.TrainId,
			&train.TrainConfig.TrainDatasetUrl,
			&train.TrainConfig.ValidDatasetUrl,
			&train.TrainConfig.DatasetShuffle,
			&train.TrainConfig.DatasetLabel,
			&train.TrainConfig.DatasetNormalizationUsage,
			&train.TrainConfig.DatasetNormalizationMethod,
			&train.TrainConfig.ModelContent,
			&train.TrainConfig.ModelConfig,
			&train.TrainConfig.CreateTime,
			&train.TrainConfig.UpdateTime,
		)
		if err != nil {
			return nil, err
		}
		trainList = append(trainList, train)
	}

	return trainList, nil
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
