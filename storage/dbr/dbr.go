package dbr

import (
	"context"
	"database/sql"

	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/go-sql-driver/mysql"
)

// DefaultDriverName is MySQL
const DefaultDriverName = DriverNameMySQL

// Connection is a connection to the database with an EventReceiver to send
// events, errors, and timings to
type Connection struct {
	DB *sql.DB
	log.Logger
	// dn internal driver name
	dn string
	// dsn Data Source Name
	dsn *mysql.Config
	// DatabaseName contains the database name to which this connection has been
	// bound to. It will only be set when a DSN has been parsed.
	DatabaseName string
}

// Session represents a business unit of execution for some connection
type Session struct {
	cxn *Connection
	log.Logger
}

type wrapContext struct {
	context.Context
	pc interface {
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	}
	qc interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}
	ec interface {
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}
	qrc interface {
		QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	}
}

// Preparer defines the only needed function to create a new prepared statement
// in the database.
type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

func (wc wrapContext) Prepare(query string) (*sql.Stmt, error) {
	return wc.pc.PrepareContext(wc.Context, query)
}

// WrapPrepareContext wraps a context around the PrepareContext function and
// returns an Preparer including your context. The provided context is used for
// the preparation of the statement, not for the execution of the statement.
func WrapPrepareContext(ctx context.Context, db interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}) Preparer {
	return wrapContext{
		Context: ctx,
		pc:      db,
	}
}

// Querier can execute a SELECT query which can return many rows.
type Querier interface {
	// Query executes a query that returns rows, typically a SELECT. The
	// args are for any placeholder parameters in the query.
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func (wc wrapContext) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return wc.qc.QueryContext(wc.Context, query, args...)
}

// WrapQueryContext wraps a context around the QueryContext function and returns
// an Querier including your context.
func WrapQueryContext(ctx context.Context, db interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}) Querier {
	return wrapContext{
		Context: ctx,
		qc:      db,
	}
}

// Execer can execute all other queries except SELECT.
type Execer interface {
	// Exec executes a query that doesn't return rows. For example: an
	// INSERT, UPDATE or DELETE or CREATE.
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (wc wrapContext) Exec(query string, args ...interface{}) (sql.Result, error) {
	return wc.ec.ExecContext(wc.Context, query, args...)
}

// WrapExecContext wraps a context around the ExecContext function and returns
// an Execer including your context.
func WrapExecContext(ctx context.Context, db interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}) Execer {
	return wrapContext{
		Context: ctx,
		ec:      db,
	}
}

// QueryRower executes a SELECT query which returns one row.
type QueryRower interface {
	// QueryRow executes a query that is expected to return at most one
	// row. QueryRow always returns a non-nil value. Errors are deferred
	// until Row's Scan method is called.
	QueryRow(query string, args ...interface{}) *sql.Row
}

func (wc wrapContext) QueryRow(query string, args ...interface{}) *sql.Row {
	return wc.qrc.QueryRowContext(wc.Context, query, args...)
}

// WrapQueryRowContext wraps a context around the QueryRowContext function and
// returns a QueryRower including your context.
func WrapQueryRowContext(ctx context.Context, db interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}) QueryRower {
	return wrapContext{
		Context: ctx,
		qrc:     db,
	}
}

// ConnectionOption can be used as an argument in NewConnection to configure a
// connection.
type ConnectionOption func(*Connection) error

// WithDB sets the DB value to a connection. If set ignores the DSN values.
func WithDB(db *sql.DB) ConnectionOption {
	return func(c *Connection) error {
		c.DB = db
		return nil
	}
}

// WithDSN sets the data source name for a connection.
func WithDSN(dsn string) ConnectionOption {
	return func(c *Connection) error {
		myc, err := mysql.ParseDSN(dsn)
		if err != nil {
			return errors.Wrap(err, "[dbr] mysql.ParseDSN")
		}
		c.dsn = myc
		return nil
	}
}

// NewConnection instantiates a Connection for a given database/sql connection
// and event receiver. An invalid drivername causes a NotImplemented error to be
// returned. You can either apply a DSN or a pre configured *sql.DB type.
func NewConnection(opts ...ConnectionOption) (*Connection, error) {
	c := &Connection{
		dn:     DriverNameMySQL,
		Logger: log.BlackHole{},
	}
	if err := c.Options(opts...); err != nil {
		return nil, errors.Wrap(err, "[dbr] NewConnection.ApplyOpts")
	}

	switch c.dn {
	case DriverNameMySQL:
	default:
		return nil, errors.NewNotImplementedf("[dbr] unsupported driver: %q", c.dn)
	}

	if c.dsn != nil {
		c.DatabaseName = c.dsn.DBName
	}

	if c.DB != nil || c.dsn == nil {
		return c, nil
	}

	var err error
	if c.DB, err = sql.Open(c.dn, c.dsn.FormatDSN()); err != nil {
		return nil, errors.Wrap(err, "[dbr] sql.Open")
	}

	return c, nil
}

// MustConnectAndVerify is like NewConnection but it verifies the connection
// and panics on errors.
func MustConnectAndVerify(opts ...ConnectionOption) *Connection {
	c, err := NewConnection(opts...)
	if err != nil {
		panic(err)
	}
	if err := c.Ping(); err != nil {
		panic(err)
	}
	return c
}

// Options applies options to a connection
func (c *Connection) Options(opts ...ConnectionOption) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return errors.Wrap(err, "[dbr] Connection ApplyOpts")
		}
	}
	return nil
}

// NewSession instantiates a Session for the Connection
func (c *Connection) NewSession(opts ...SessionOption) *Session {
	s := &Session{
		cxn:    c,
		Logger: c.Logger,
	}
	s.Options(opts...)
	return s
}

// Close closes the database, releasing any open resources.
func (c *Connection) Close() error {
	return errors.Wrap(c.DB.Close(), "[dbr] connection.close")
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
func (c *Connection) Ping() error {
	return errors.Wrap(c.DB.Ping(), "[dbr] connection.ping")
}

// SessionOption can be used as an argument in NewSession to configure a session.
type SessionOption func(cxn *Connection, s *Session) error

// Options applies options to a session
func (s *Session) Options(opts ...SessionOption) error {
	for _, opt := range opts {
		if err := opt(s.cxn, s); err != nil {
			return errors.Wrap(err, "[dbr] Session.Options")
		}
	}
	return nil
}

// SessionRunner can do anything that a Session can except start a transaction.
//type SessionRunner interface {
//	Select(cols ...string) *Select
//	SelectBySql(sql string, args ...interface{}) *Select
//
//	InsertInto(into string) *Insert
//	Update(table ...string) *Update
//	UpdateBySql(sql string, args ...interface{}) *Update
//	DeleteFrom(from ...string) *Delete
//}
