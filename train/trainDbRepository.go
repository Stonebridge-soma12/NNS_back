package train

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
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

func (tdb *TrainDbRepository) Insert(train Train) error {
	options := options{}
	ApplyOptions(&options, insertTrain())

	_, err := tdb.DB.NamedExec(options.queryString, &train)
	if err != nil {
		return err
	}

	return nil
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
	fmt.Println(options.queryString)
	fmt.Println(options.args)
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
