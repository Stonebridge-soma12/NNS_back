package train

import (
	"github.com/elixter/Querybuilder"
	"github.com/jmoiron/sqlx"
)

const (
	defaultSelectEpochQuery = "SELECT e.id, train_id, epoch, acc, loss, val_acc, val_loss, learning_rate, create_time, update_time FROM epoch e "
	defaultSelectEpochColumns = "e.id, e.train_id, e.epoch, e.acc, e.loss, e.val_acc, e.val_loss, e.learning_rate, e.create_time, e.update_time"
)


func WithEpochTrainId(trainId int64) query.Option {
	return query.OptionFunc(func(b *query.Builder) {
		b.AddWhere("e.id = ?", trainId)
	})
}

type EpochDbRepository struct {
	DB *sqlx.DB
}

func (edr *EpochDbRepository) Insert(epoch Epoch) error {
	builder := query.Builder{}
	builder.AddInsert(
		"epoch",
		"train_id, epoch, acc, loss, val_acc, val_loss, learning_rate",
		":train_id, :epoch, :acc, :loss, :val_acc, :val_loss, :learning_rate",
	)

	err := builder.Build()
	if err != nil {
		return err
	}

	_, err = edr.DB.NamedExec(builder.QueryString, &epoch)
	if err != nil {
		return err
	}

	return nil
}

func (edr *EpochDbRepository) Find(opts ...query.Option) (Epoch, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectEpochColumns).
		AddFrom("epoch e").
		AddJoin("train t ON e.train_id = t.id").
		AddJoin("project p ON t.project_id = p.id")

	var epoch Epoch

	err := builder.Build()
	if err != nil {
		return epoch, err
	}

	err = edr.DB.Get(&epoch, builder.QueryString, builder.Args...)
	if err != nil {
		return epoch, err
	}

	return epoch, nil
}

func (edr *EpochDbRepository) FindAll(opts ...query.Option) ([]Epoch, error) {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddSelect(defaultSelectEpochColumns).
		AddFrom("epoch e").
		AddJoin("train t ON e.train_id = t.id").
		AddJoin("project p ON t.project_id = p.id")

	err := builder.Build()
	if err != nil {
		return nil, err
	}

	var epochs []Epoch
	rows, err := edr.DB.Queryx(builder.QueryString, builder.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var epoch Epoch
		err := rows.StructScan(&epoch)
		if err != nil {
			return nil, err
		}
		epochs = append(epochs, epoch)
	}

	return epochs, nil
}

func (edr *EpochDbRepository) Delete(opts ...query.Option) error {
	builder := query.ApplyQueryOptions(opts...)
	builder.AddDelete().
		AddFrom("epoch e")

	err := builder.Build()
	if err != nil {
		return err
	}

	_, err = edr.DB.Exec(builder.QueryString,builder.Args)
	if err != nil {
		return err
	}

	return nil
}