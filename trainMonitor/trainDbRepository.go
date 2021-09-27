package trainMonitor

import "github.com/jmoiron/sqlx"

const (
	defaultSelectTrainQuery = "select * from train "
	defaultDeleteTrainQuery = "delete from train "
)

type TrainDbRepository struct {
	DB *sqlx.DB
}

func insertTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into " +
			"train (status, acc, loss, val_acc, val_loss, epochs, name, url) " +
			"values(:status, :acc, :loss, :val_acc, :val_loss, :epochs, :name, :url)"
	})
}

func updateTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "update train " +
			"set status=:status, acc=:acc, loss=:loss, val_acc=:val_acc, val_loss=:val_loss, epochs=:epochs, name=:name " +
			"where id = :id"
	})
}

func WithUserIdAndProjectNo(UserId int64, TrainNo int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where user_id = ? and train_id = ?"
		o.args = append(o.args, UserId)
		o.args = append(o.args, TrainNo)
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

func (tdb *TrainDbRepository) Find(opts ... Option) (Train, error) {
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
	options := options {
		queryString: defaultSelectTrainQuery,
	}
	ApplyOptions(&options, opts...)

	var trains []Train
	rows, err := tdb.DB.Queryx(options.queryString, options.args)
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