package query

type Option interface {
	apply(*Builder)
}

type OptionFunc func(*Builder)

func (o OptionFunc) apply(q *Builder) {
	o(q)
}

func ApplyQueryOptions(opts ...Option) *Builder {
	result := new(Builder)

	for _, o := range opts {
		o.apply(result)
	}

	return result
}
