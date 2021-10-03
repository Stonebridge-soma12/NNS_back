package train

import (
	"github.com/jmoiron/sqlx"
)

const (
	defaultSelectTrainQuery = `
SELECT id,
       user_id,
       train_no,
       project_id,
       acc,
       loss,
       val_acc,
       val_loss,
       name,
       epochs,
       url,
       status
FROM train 
`
	defaultDeleteTrainQuery = "delete from train "
)

type TrainDbRepository struct {
	DB *sqlx.DB
}

func (tdb *TrainDbRepository) FindNextTrainNo(userId int64) (int64, error) {
	panic("implement me")
}

func insertTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = `
INSERT INTO train (user_id,
                   train_no,
                   project_id,
                   acc,
                   loss,
                   val_acc,
                   val_loss,
                   name,
                   epochs,
                   url,
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
        :url,
        :status)
`
	})
}

func updateTrain() Option {
	return optionFunc(func(o *options) {
		o.queryString = "update train " +
			"set status=:status, acc=:acc, loss=:loss, val_acc=:val_acc, val_loss=:val_loss, epochs=:epochs, name=:name, url=:url " +
			"where id = :id"
	})
}

func WithUserId(userId int64) Option {
	return optionFunc(func(o *options) {
		o.queryString += `WHERE user_id = ? `
		o.args = append(o.args, userId)
	})
}

func WithUserIdAndProjectNo(userId int64, projectNo int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where project_id in " +
			"(select id " +
			"from project " +
			"where user_id = ? and project_no = ?) "
		o.args = append(o.args, userId)
		o.args = append(o.args, projectNo)
	})
}

func WithUserIdAndProjectNoAndTrainNo(userId int64, projectNo int, trainNo int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where train_no = ? and project_id in " +
			"(select id " +
			"from project " +
			"where user_id = ? and project_no = ?) "
		o.args = append(o.args, trainNo)
		o.args = append(o.args, userId)
		o.args = append(o.args, projectNo)
	})
}

func WithLimit(offset, limit int) Option {
	return optionFunc(func(o *options) {
		o.queryString += `LIMIT ?, ? `
		o.args = append(o.args, offset, limit)
	})
}

func (tdb *TrainDbRepository) Insert(train Train) (int64, error) {
	options := options{}
	ApplyOptions(&options, insertTrain())

	result, err := tdb.DB.NamedExec(options.queryString, &train)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
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