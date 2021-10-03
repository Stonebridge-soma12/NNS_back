package train

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	defaultSelectTrainQuery = "select * from train "
	defaultDeleteTrainQuery = "update train set status='DEL' "
	defaultSelectTrain      = `select t.id,
									   t.status,
									   t.acc,
									   t.loss,
									   t.val_acc,
									   t.val_loss,
									   t.name,
									   t.epochs,
									   t.project_id,
									   t.url,
									   t.train_no
								from train t
								join project p on t.project_id = p.id`
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
			"set status=:status, acc=:acc, loss=:loss, val_acc=:val_acc, val_loss=:val_loss, epochs=:epochs, name=:name, url=:url " +
			"where id = :id"
	})
}

func WithUserIdAndProjectNo(userId int64, projectNo int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where project_id in " +
			"(select id " +
			"from project " +
			"where user_id = ? and project_no = ?);"
		o.args = append(o.args, userId)
		o.args = append(o.args, projectNo)
	})
}

func WithUserIdAndProjectNoAndTrainNo(userId int64, projectNo int, trainNo int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where train_no = ? and project_id in " +
			"(select id " +
			"from project " +
			"where user_id = ? and project_no = ?);"
		o.args = append(o.args, trainNo)
		o.args = append(o.args, userId)
		o.args = append(o.args, projectNo)
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
