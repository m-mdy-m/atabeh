package core

import "strings"

type Query struct {
	selects []string
	from    string
	where   []string
	order   string
	limit   string
	args    []any
}

func Select(fields ...string) *Query {
	return &Query{selects: fields}
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

func (q *Query) OrderBy(o string) *Query {
	q.order = o
	return q
}

func (q *Query) Limit(n int) *Query {
	q.limit = "LIMIT ?"
	q.args = append(q.args, n)
	return q
}

func (q *Query) And(cond string, args ...any) *Query {
	return q.Where(cond, args...)
}

func (q *Query) Or(cond string, args ...any) *Query {
	if len(q.where) == 0 {
		return q.Where(cond, args...)
	}
	last := q.where[len(q.where)-1]
	q.where[len(q.where)-1] = "(" + last + " OR " + cond + ")"
	q.args = append(q.args, args...)
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
func (q *Query) Build() (string, []any) {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	sb.WriteString(strings.Join(q.selects, ", "))
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

	return sb.String(), q.args
}
