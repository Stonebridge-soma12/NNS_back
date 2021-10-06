package query

import "fmt"

type Builder struct {
	action        string
	update        []string
	updateTable   string
	insert        string
	insertColumns string
	insertValues  string
	selects       []string
	from          []string
	join          []string
	where         []string
	order         []string
	limit         string
	updateArgs    []interface{}
	joinArgs      []interface{}
	whereArgs     []interface{}
	limitArgs     []interface{}
	Args          []interface{}
	QueryString   string
}

const (
	actionDelete = "DEL"
	actionSelect = "SEL"
	actionUpdate = "UP"
	actionAlter  = "ALT"
	actionInsert = "INS"
)

const (
	ErrEmptySelect = "must specified selecting columns"
	ErrEmptyFrom   = "must specified selecting table"
)

func (b *Builder) AddDelete() *Builder {
	b.action = actionDelete

	return b
}

func (b *Builder) AddUpdate(table string, set string, args ...interface{}) *Builder {
	b.action = actionUpdate
	b.updateTable = table
	b.update = append(b.update, set)
	b.updateArgs = append(b.updateArgs, args...)

	return b
}

func (b *Builder) AddInsert(table, columns, values string) *Builder {
	b.action = actionInsert
	b.insert = table
	b.insertColumns = columns
	b.insertValues = values

	return b
}

func (b *Builder) AddSelect(columns string) *Builder {
	b.action = actionSelect
	b.selects = append(b.selects, columns)

	return b
}

func (b *Builder) AddFrom(table string) *Builder {
	b.from = append(b.from, table)

	return b
}

func (b *Builder) AddJoin(join string, args ...interface{}) *Builder {
	b.join = append(b.join, join)
	b.joinArgs = append(b.joinArgs, args...)

	return b
}

func (b *Builder) AddWhere(where string, args ...interface{}) *Builder {
	b.where = append(b.where, where)
	b.whereArgs = append(b.whereArgs, args...)

	return b
}

func (b *Builder) AddOrder(order string) *Builder {
	b.order = append(b.order, order)

	return b
}

func (b *Builder) AddLimit(offset, limit int) *Builder {
	b.limit = "LIMIT ?, ?"
	b.limitArgs = append(b.limitArgs, offset, limit)

	return b
}

func (b *Builder) Build() error {
	b.QueryString = ""

	switch b.action {
	case actionDelete:
		b.QueryString += "DELETE "
	case actionSelect:
		if len(b.selects) <= 0 {
			return fmt.Errorf(ErrEmptySelect)
		}
		b.QueryString += "SELECT "
		for _, cols := range b.selects {
			b.QueryString += cols + " "
		}
		if len(b.from) <= 0 {
			return fmt.Errorf(ErrEmptyFrom)
		}
	case actionUpdate:
		b.QueryString += "UPDATE " + b.updateTable + " "
	case actionInsert:
		b.QueryString += fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) ", b.insert, b.insertColumns, b.insertValues)
	}

	if len(b.from) > 0 {
		b.QueryString += "FROM "
		for _, cols := range b.from {
			b.QueryString += cols + " "
		}
	}

	if len(b.joinArgs) > 0 {
		b.Args = append(b.Args, b.joinArgs...)
	}
	for _, cols := range b.join {
		b.QueryString += "JOIN " + cols + " "
	}

	if len(b.update) > 0 {
		b.QueryString += "SET "
	}
	for i, cols := range b.update {
		b.QueryString += cols
		if i < len(b.update)-1 {
			b.QueryString += ", "
		} else {
			b.QueryString += " "
		}
	}

	if len(b.updateArgs) > 0 {
		b.Args = append(b.Args, b.updateArgs...)
	}

	if len(b.where) > 0 {
		b.QueryString += "WHERE "

		if len(b.whereArgs) > 0 {
			b.Args = append(b.Args, b.whereArgs...)
		}
	}
	for i, cols := range b.where {
		if i != 0 {
			b.QueryString += "AND "
		}
		b.QueryString += cols + " "
	}

	if len(b.order) > 0 {
		b.QueryString += "ORDER BY "
	}
	for i, cols := range b.order {
		b.QueryString += cols
		if i != len(b.where)-1 {
			b.QueryString += ", "
		} else {
			b.QueryString += " "
		}
	}

	if b.limit != "" {
		b.Args = append(b.Args, b.limitArgs...)
	}
	b.QueryString += b.limit

	return nil
}
