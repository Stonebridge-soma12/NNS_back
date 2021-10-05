package query

import "fmt"

type Query struct {
	selects     []string
	from        []string
	join        []string
	where       []string
	order       []string
	limit       string
	joinArgs    []interface{}
	whereArgs   []interface{}
	limitArgs   []interface{}
	Args        []interface{}
	QueryString string
}

const (
	ErrEmptySelect = "must specified selecting columns"
	ErrEmptyFrom   = "must specified selecting table"
)

func (q *Query) AddSelect(columns string) *Query {
	q.selects = append(q.selects, columns)

	return q
}

func (q *Query) AddFrom(table string) *Query {
	q.from = append(q.from, table)

	return q
}

func (q *Query) AddJoin(join string, args ...interface{}) *Query {
	q.join = append(q.join, join)
	q.joinArgs = append(q.joinArgs, args...)

	return q
}

func (q *Query) AddWhere(where string, args ...interface{}) *Query {
	q.where = append(q.where, where)
	q.whereArgs = append(q.whereArgs, args...)

	return q
}

func (q *Query) AddOrder(order string) *Query {
	q.order = append(q.order, order)

	return q
}

func (q *Query) AddLimit(offset, limit int) *Query {
	q.limit = "LIMIT ?, ?"
	q.limitArgs = append(q.limitArgs, offset, limit)

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

	if len(q.join) > 0 {
		q.Args = append(q.Args, q.joinArgs)
	}
	for _, cols := range q.join {
		q.QueryString += "JOIN " + cols + " "
	}

	if len(q.where) > 0 {
		q.QueryString += "WHERE "
		q.Args = append(q.Args, q.whereArgs)
	}
	for i, cols := range q.where {
		if i != 0 {
			q.QueryString += "AND "
		}
		q.QueryString += cols + " "
	}

	if len(q.order) > 0 {
		q.QueryString += "ORDER BY "
	}
	for i, cols := range q.order {
		q.QueryString += cols
		if i != len(q.where) - 1 {
			q.QueryString += ", "
		} else {
			q.QueryString += " "
		}
	}

	if q.limit != "" {
		q.Args = append(q.Args, q.limitArgs)
	}
	q.QueryString += q.limit

	return nil
}