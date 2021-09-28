package train

import "github.com/jmoiron/sqlx"

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
	options := options{}
	insertEpoch().apply(&options)

	_, err := edr.DB.NamedExec(options.queryString, &epoch)
	if err != nil {
		return err
	}

	return nil
}

func (edr *EpochDbRepository) Find(opts ...Option) (Epoch, error) {
	options := options{
		queryString: defaultSelectEpochQuery,
	}

	ApplyOptions(&options, opts...)

	var epoch Epoch
	err := edr.DB.Get(&epoch, options.queryString, options.args...)
	if err != nil {
		return epoch, err
	}

	return epoch, nil
}

func (edr *EpochDbRepository) FindAll(opts ...Option) ([]Epoch, error) {
	options := options{
		queryString: defaultSelectEpochQuery,
	}

	ApplyOptions(&options, opts...)

	var epochs []Epoch
	rows, err := edr.DB.Queryx(options.queryString, options.args)
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