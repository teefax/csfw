package dbr

import (
	"database/sql"
	"database/sql/driver"

	"strconv"

	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/util/bufferpool"
	"github.com/corestoreio/csfw/util/errors"
)

type expr struct {
	SQL    string
	Values []interface{}
}

// Expr is a SQL fragment with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *expr {
	return &expr{SQL: sql, Values: values}
}

// Update contains the clauses for an UPDATE statement
type Update struct {
	log.Logger
	Execer
	Preparer

	RawFullSQL   string
	RawArguments []interface{}

	Table          alias
	SetClauses     []*setClause
	WhereFragments []*whereFragment
	OrderBys       []string
	LimitCount     uint64
	LimitValid     bool
	OffsetCount    uint64
	OffsetValid    bool
}

type setClause struct {
	column string
	value  interface{}
}

// Update creates a new Update for the given table
func (sess *Session) Update(table ...string) *Update {
	return &Update{
		Logger: sess.Logger,
		Execer: sess.cxn.DB,
		Table:  MakeAlias(table...),
	}
}

// UpdateBySql creates a new Update for the given SQL string and arguments
func (sess *Session) UpdateBySql(sql string, args ...interface{}) *Update {
	if err := argsValuer(&args); err != nil {
		//sess.EventErrKv("dbr.insertbuilder.values", err, kvs{"args": fmt.Sprint(args)})
		panic(err) // todo remove panic
	}
	return &Update{
		Logger:       sess.Logger,
		Execer:       sess.cxn.DB,
		RawFullSQL:   sql,
		RawArguments: args,
	}
}

// Update creates a new Update for the given table bound to a transaction
func (tx *Tx) Update(table ...string) *Update {
	return &Update{
		Logger: tx.Logger,
		Execer: tx.Tx,
		Table:  MakeAlias(table...),
	}
}

// UpdateBySql creates a new Update for the given SQL string and arguments bound to a transaction
func (tx *Tx) UpdateBySql(sql string, args ...interface{}) *Update {
	if err := argsValuer(&args); err != nil {
		// tx.EventErrKv("dbr.insertbuilder.values", err, kvs{"args": fmt.Sprint(args)})
		panic(err) // todo remove panic
	}
	return &Update{
		Logger:       tx.Logger,
		Execer:       tx.Tx,
		RawFullSQL:   sql,
		RawArguments: args,
	}
}

// Set appends a column/value pair for the statement
func (b *Update) Set(column string, value interface{}) *Update {
	if dbVal, ok := value.(driver.Valuer); ok {
		if val, err := dbVal.Value(); err == nil {
			value = val
		} else {
			panic(err)
		}
	}
	b.SetClauses = append(b.SetClauses, &setClause{column: column, value: value})
	return b
}

// SetMap appends the elements of the map as column/value pairs for the statement
func (b *Update) SetMap(clauses map[string]interface{}) *Update {
	for col, val := range clauses {
		b = b.Set(col, val)
	}
	return b
}

// Where appends a WHERE clause to the statement
func (b *Update) Where(args ...ConditionArg) *Update {
	b.WhereFragments = append(b.WhereFragments, newWhereFragments(args...)...)
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *Update) OrderBy(ord string) *Update {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

// OrderDir appends a column to ORDER the statement by with a given direction
func (b *Update) OrderDir(ord string, isAsc bool) *Update {
	if isAsc {
		b.OrderBys = append(b.OrderBys, ord+" ASC")
	} else {
		b.OrderBys = append(b.OrderBys, ord+" DESC")
	}
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *Update) Limit(limit uint64) *Update {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *Update) Offset(offset uint64) *Update {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

// ToSQL serialized the Update to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *Update) ToSQL() (string, []interface{}, error) {
	if b.RawFullSQL != "" {
		return b.RawFullSQL, b.RawArguments, nil
	}

	if len(b.Table.Expression) == 0 {
		return "", nil, errors.NewEmptyf("[dbr] Update: Table is empty")
	}
	if len(b.SetClauses) == 0 {
		return "", nil, errors.NewEmptyf("[dbr] Update: SetClauses are empty")
	}

	var buf = bufferpool.Get()
	defer bufferpool.Put(buf)

	var args []interface{}

	buf.WriteString("UPDATE ")
	buf.WriteString(b.Table.QuoteAs())
	buf.WriteString(" SET ")

	// Build SET clause SQL with placeholders and add values to args
	for i, c := range b.SetClauses {
		if i > 0 {
			buf.WriteString(", ")
		}
		Quoter.writeQuotedColumn(c.column, buf)
		if e, ok := c.value.(*expr); ok {
			buf.WriteString(" = ")
			buf.WriteString(e.SQL)
			args = append(args, e.Values...)
		} else {
			buf.WriteString(" = ?")
			args = append(args, c.value)
		}
	}

	// Write WHERE clause if we have any fragments
	if len(b.WhereFragments) > 0 {
		buf.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.WhereFragments, buf, &args)
	}

	// Ordering and limiting
	if len(b.OrderBys) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, s := range b.OrderBys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(s)
		}
	}

	if b.LimitValid {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.FormatUint(b.LimitCount, 10))
	}

	if b.OffsetValid {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.FormatUint(b.OffsetCount, 10))
	}

	return buf.String(), args, nil
}

// Exec executes the statement represented by the Update
// It returns the raw database/sql Result and an error if there was one
func (b *Update) Exec() (sql.Result, error) {
	rawSQL, args, err := b.ToSQL()
	if err != nil {
		return nil, errors.Wrap(err, "[dbr] Update.Exec.ToSQL")
	}

	fullSql, err := Preprocess(rawSQL, args)
	if err != nil {
		return nil, errors.Wrap(err, "[dbr] Update.Exec.Preprocess")
	}

	if b.Logger != nil && b.Logger.IsInfo() {
		defer log.WhenDone(b.Logger).Info("dbr.Update.Exec.Timing", log.String("sql", fullSql))
	}

	result, err := b.Execer.Exec(fullSql)
	if err != nil {
		return result, errors.Wrap(err, "[dbr] Update.Exec.Exec")
	}

	return result, nil
}

// Exec executes the statement represented by the Update
// It returns the raw database/sql Result and an error if there was one
func (b *Update) Prepare() (*sql.Stmt, error) {
	rawSQL, _, err := b.ToSQL() // TODO create a ToSQL version without any arguments
	if err != nil {
		return nil, errors.Wrap(err, "[dbr] Update.Prepare.ToSQL")
	}

	if b.Logger != nil && b.Logger.IsInfo() {
		defer log.WhenDone(b.Logger).Info("dbr.Update.Prepare.Timing", log.String("sql", rawSQL))
	}

	stmt, err := b.Preparer.Prepare(rawSQL)
	return stmt, errors.Wrap(err, "[dbr] Update.Prepare.Prepare")
}
