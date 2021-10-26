package train

import (
	"github.com/elixter/Querybuilder"
	"github.com/jmoiron/sqlx"
)

const (
	defaultSelectTrainLogQuery = "SELECT id, train_id, msg, status_code, create_time, update_time FROM train_log "
	defaultDeleteTrainLogQuery = "DELETE FROM train_log "
	defaultSelectTrainLogColumns = "id, train_id, msg, status_code, create_time, update_time"
)

type TrainLogDbRepository struct {
	DB *sqlx.DB
}

func insertLog() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into trainLog (train_id, msg) " +
			"values (:train_id, :msg)"
	})
}

func (ldr *TrainLogDbRepository) Insert(trainLog TrainLog) error {
	builder := query.Builder{}
	builder.AddInsert(
			"train_log",
			"train_id, msg",
			":train_id, :msg",
		)

	err := builder.Build()
	if err != nil {
		return err
	}

	_, err = ldr.DB.NamedExec(builder.QueryString, &trainLog)
	if err != nil {
		return err
	}

	return nil
}

func (ldr *TrainLogDbRepository) Delete(opts ...query.Option) error {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddDelete().
		AddFrom("train_log")

	_, err := ldr.DB.Exec(builder.QueryString, builder.Args)
	if err != nil {
		return err
	}

	return nil
}

func (ldr *TrainLogDbRepository) Find(opts ...query.Option) (TrainLog, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectTrainLogColumns).
		AddFrom("train_log")

	err := builder.Build()
	if err != nil {
		return TrainLog{}, err
	}

	var trainLog TrainLog
	err = ldr.DB.Get(&trainLog, builder.QueryString, builder.Args...)
	if err != nil {
		return TrainLog{}, err
	}

	return trainLog, err
}

func (ldr *TrainLogDbRepository) FindAll(opts ... query.Option) ([]TrainLog, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectTrainLogColumns).
		AddFrom("train_log")

	err := builder.Build()
	if err != nil {
		return nil, err
	}

	rows, err := ldr.DB.Queryx(builder.QueryString, builder.Args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []TrainLog
	for rows.Next() {
		var log TrainLog
		err := rows.StructScan(&log)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}