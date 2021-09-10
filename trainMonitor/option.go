package trainMonitor

type options struct {
	queryString string
	args        []interface{}
	queryId     int // id of selected record
}

type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

func ApplyOptions(target *options, opts ...Option) {
	for _, o := range opts {
		o.apply(target)
	}
}

func WithTrainID(trainId int) Option {
	return optionFunc(func(o *options) {
		o.queryString += "where train_id = ?"
		o.args = append(o.args, trainId)
	})
}

func WithId(id int) Option {
	return optionFunc(func (o *options) {
		o.queryString += "where id = ?"
		o.args = append(o.args, id)
	})
}