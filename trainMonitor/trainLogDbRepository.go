package trainMonitor

import "github.com/jmoiron/sqlx"

const (
	defaultSelectTrainLogQuery = "select * from TrainLog "
	defaultDeleteTrainLogQuery = "delete from TrainLog "
)

type TrainLogDbRepository struct {
	DB *sqlx.DB
}

func insertLog() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into TrainLog (train_id, msg) " +
			"values (:train_id, :msg)"
	})
}

func (ldr *TrainLogDbRepository) Insert(trainLog TrainLog) error {
	options := options{}
	insertLog().apply(&options)

	_, err := ldr.DB.NamedExec(options.queryString, &trainLog)
	if err != nil {
		return err
	}

	return nil
}

func (ldr *TrainLogDbRepository) Delete(opts ...Option) error {
	options := options {
		queryString: defaultDeleteTrainLogQuery,
	}
	ApplyOptions(&options, opts...)

	_, err := ldr.DB.Exec(options.queryString, options.args)
	if err != nil {
		return err
	}

	return nil
}

func (ldr *TrainLogDbRepository) Find(opts ...Option) (TrainLog, error) {
	options := options{
		queryString: defaultSelectTrainLogQuery,
	}
	ApplyOptions(&options, opts...)

	var trainLog TrainLog
	err := ldr.DB.Get(&trainLog, options.queryString, options.args)
	if err != nil {
		return TrainLog{}, err
	}

	return trainLog, err
}

func (ldr *TrainLogDbRepository) FindAll(opts ... Option) ([]TrainLog, error) {
	options := options{
		queryString: defaultSelectEpochQuery,
	}
	ApplyOptions(&options, opts...)

	rows, err := ldr.DB.Queryx(options.queryString, options.args)
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