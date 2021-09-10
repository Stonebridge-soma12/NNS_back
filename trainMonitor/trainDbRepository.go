package trainMonitor

import "github.com/jmoiron/sqlx"

const (
	defaultSelectTrainQuery = "select * from Train "
	defaultDeleteTrainQuery = "delete from Train "
)

type TrainDbRepository struct {
	DB *sqlx.DB
}

func insertTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into " +
			"Train (status, acc, loss, val_acc, val_loss, epochs, name) " +
			"values(:status, :acc, :loss, :val_acc, :val_loss, :epochs, :name)"
	})
}

func updateTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "update Train " +
			"set status=:status, acc=:acc, loss=:loss, val_acc=:val_acc, val_loss=:val_loss, epochs=:epochs, name=:name " +
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

func (tdb *TrainDbRepository) Find(opts ... Option) (Train, error) {
	options := options{
		queryString: defaultSelectTrainQuery,
	}
	ApplyOptions(&options, opts...)

	var train Train
	err := tdb.DB.Get(&train, options.queryString, options.args)
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