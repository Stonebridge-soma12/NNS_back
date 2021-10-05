package query

type Option interface {
	apply(*Query)
}

type OptionFunc func(*Query)

func (o OptionFunc) apply(q *Query) {
	o(q)
}

func ApplyQueryOptions(opts ...Option) *Query {
	var result *Query

	for _, o := range opts {
		o.apply(result)
	}

	return result
}
