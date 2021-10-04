package train

import "fmt"

type Query struct {
	selects     []string
	from        []string
	join        []string
	where       []string
	limit       string
	Args        []interface{}
	QueryString string
}

const (
	ErrEmptySelect = "must specified selecting columns"
	ErrEmptyFrom   = "must specified selecting table"
)

func (q *Query) AddSelect(columns string, args ...interface{}) *Query {
	q.selects = append(q.selects, columns)
	q.Args = append(q.Args, args...)

	return q
}

func (q *Query) AddFrom(table string, args ...interface{}) *Query {
	q.from = append(q.from, table)
	q.Args = append(q.Args, args...)

	return q
}

func (q *Query) AddJoin(join string, args ...interface{}) *Query {
	q.join = append(q.join, join)
	q.Args = append(q.Args, args...)

	return q
}

func (q *Query) AddWhere(where string, args ...interface{}) *Query {
	q.where = append(q.where, where)
	q.Args = append(q.Args, args...)

	return q
}

func (q *Query) AddLimit(offset, limit int) *Query {
	q.limit = "LIMIT ?, ?"
	q.Args = append(q.Args, offset, limit)

	return q
}

func (q *Query) Apply() error {
	q.QueryString = ""
	if len(q.selects) <= 0 {
		return fmt.Errorf(ErrEmptySelect)
	}

	if len(q.from) <= 0 {
		return fmt.Errorf(ErrEmptyFrom)
	}

	q.QueryString += "SELECT "
	for _, cols := range q.selects {
		q.QueryString += cols + " "
	}

	q.QueryString += "FROM "
	for _, cols := range q.from {
		q.QueryString += cols + " "
	}

	for _, cols := range q.join {
		q.QueryString += "JOIN " + cols + " "
	}

	if len(q.where) > 0 {
		q.QueryString += "WHERE "
	}
	for i, cols := range q.where {
		if i != 0 {
			q.QueryString += "AND "
		}
		q.QueryString += cols + " "
	}

	q.QueryString += q.limit

	return nil
}
