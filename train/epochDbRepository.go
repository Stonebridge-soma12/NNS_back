package train

import (
	"github.com/jmoiron/sqlx"
	"nns_back/query"
)

const (
	defaultSelectEpochQuery = "SELECT e.id, train_id, epoch, acc, loss, val_acc, val_loss, learning_rate, create_time, update_time FROM epoch e "
	defaultDeleteEpochQuery = "DELETE FROM epoch "

	defaultSelectEpochcolumns = "e.id, train_id, epoch, acc, loss, val_acc, val_loss, learning_rate, create_time, update_time"
)


type EpochDbRepository struct {
	DB *sqlx.DB
}

func insertEpoch() Option {
	return optionFunc(func(o *options) {
		o.queryString = "insert into " +
			"epoch(train_id, epoch, acc, loss, val_acc, val_loss, learning_rate) " +
			"values (:train_id, :epoch, :acc, :loss, :val_acc, :val_loss, :learning_rate)"
	})
}

func (edr *EpochDbRepository) Insert(epoch Epoch) error {
	builder := query.Builder{}
	builder.AddInsert(
		"epoch(train_id, epoch, acc, loss, val_acc, val_loss, learning_rate)",
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
	builder.AddSelect(defaultSelectEpochcolumns).
		AddFrom("epoch e")

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
	builder.AddSelect(defaultSelectEpochcolumns).
		AddFrom("epoch e")

	err := builder.Build()
	if err != nil {
		return nil, err
	}

	var epochs []Epoch
	rows, err := edr.DB.Queryx(builder.QueryString, builder.Args)
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

func (edr *EpochDbRepository) Delete(opts ...Option) error {
	options := options{
		queryString: defaultDeleteEpochQuery,
	}

	ApplyOptions(&options, opts...)

	_, err := edr.DB.Exec(options.queryString, options.args)
	if err != nil {
		return err
	}

	return nil
}