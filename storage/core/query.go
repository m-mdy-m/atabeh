package core

import "strings"

type Query struct {
	kind    string
	selects []string
	from    string
	where   []string
	order   string
	limit   string

	insertTable    string
	insertCols     []string
	insertVals     []any
	insertModifier string
	updateTable    string
	sets           []string
	args           []any
}

func Select(fields ...string) *Query {
	return &Query{kind: "select", selects: fields}
}

func InsertInto(table string) *Query {
	return &Query{kind: "insert", insertTable: table}
}

func Update(table string) *Query {
	return &Query{kind: "update", updateTable: table}
}

func DeleteFrom(table string) *Query {
	return &Query{kind: "delete", from: table}
}

func (q *Query) From(table string) *Query {
	q.from = table
	return q
}

func (q *Query) Where(cond string, args ...any) *Query {
	q.where = append(q.where, cond)
	q.args = append(q.args, args...)
	return q
}

func (q *Query) And(cond string, args ...any) *Query { return q.Where(cond, args...) }

func (q *Query) Or(cond string, args ...any) *Query {
	if len(q.where) == 0 {
		return q.Where(cond, args...)
	}
	last := q.where[len(q.where)-1]
	q.where[len(q.where)-1] = "(" + last + " OR " + cond + ")"
	q.args = append(q.args, args...)
	return q
}

func (q *Query) OrderBy(o string) *Query {
	q.order = o
	return q
}

func (q *Query) Limit(n int) *Query {
	q.limit = "LIMIT ?"
	q.args = append(q.args, n)
	return q
}

func (q *Query) In(column string, values ...any) *Query {
	if len(values) == 0 {
		q.Where("1=0")
		return q
	}
	placeholders := strings.Repeat("?,", len(values))
	placeholders = placeholders[:len(placeholders)-1]
	q.Where(column+" IN ("+placeholders+")", values...)
	return q
}

func (q *Query) Columns(cols ...string) *Query {
	q.insertCols = append(q.insertCols, cols...)
	return q
}

func (q *Query) Values(vals ...any) *Query {
	q.insertVals = append(q.insertVals, vals...)
	q.args = append(q.args, vals...)
	return q
}

func (q *Query) OrIgnore() *Query {
	q.insertModifier = "OR IGNORE"
	return q
}

func (q *Query) OrReplace() *Query {
	q.insertModifier = "OR REPLACE"
	return q
}

func (q *Query) Set(column string, val any) *Query {
	q.sets = append(q.sets, column+" = ?")
	q.args = append(q.args, val)
	return q
}

func (q *Query) Build() (string, []any) {
	switch q.kind {
	case "select":
		return q.buildSelect(), q.args
	case "insert":
		return q.buildInsert(), q.args
	case "update":
		return q.buildUpdate(), q.args
	case "delete":
		return q.buildDelete(), q.args
	default:
		return "", nil
	}
}

func (q *Query) buildSelect() string {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	if len(q.selects) == 0 {
		sb.WriteString("*")
	} else {
		sb.WriteString(strings.Join(q.selects, ", "))
	}
	sb.WriteString(" FROM ")
	sb.WriteString(q.from)

	if len(q.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(q.where, " AND "))
	}
	if q.order != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(q.order)
	}
	if q.limit != "" {
		sb.WriteString(" ")
		sb.WriteString(q.limit)
	}
	return sb.String()
}

func (q *Query) buildInsert() string {
	var sb strings.Builder
	sb.WriteString("INSERT ")
	if q.insertModifier != "" {
		sb.WriteString(q.insertModifier + " ")
	}
	sb.WriteString("INTO ")
	sb.WriteString(q.insertTable)

	if len(q.insertCols) > 0 {
		sb.WriteString(" (")
		sb.WriteString(strings.Join(q.insertCols, ", "))
		sb.WriteString(")")
	}

	if len(q.insertVals) > 0 {
		n := len(q.insertVals)
		placeholders := strings.Repeat("?,", n)
		placeholders = placeholders[:len(placeholders)-1]
		sb.WriteString(" VALUES (")
		sb.WriteString(placeholders)
		sb.WriteString(")")
	} else {

		sb.WriteString(" DEFAULT VALUES")
	}
	return sb.String()
}

func (q *Query) buildUpdate() string {
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	sb.WriteString(q.updateTable)
	sb.WriteString(" SET ")
	sb.WriteString(strings.Join(q.sets, ", "))

	if len(q.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(q.where, " AND "))
	}
	return sb.String()
}

func (q *Query) buildDelete() string {
	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	sb.WriteString(q.from)
	if len(q.where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(q.where, " AND "))
	}
	return sb.String()
}
