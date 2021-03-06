// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package csdb

import (
	"context"

	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/util/errors"
)

// Table represents a table from a database.
type Table struct {
	// Schema represents the name of the database. Might be empty.
	Schema string
	// Name of the table
	Name string
	// Columns all table columns
	Columns Columns
	// CountPK number of primary keys. Auto updated.
	CountPK int
	// CountUnique number of unique keys. Auto updated.
	CountUnique int

	// internal caches
	fieldsPK  []string // all PK column field
	fieldsUNI []string // all unique key column field
	fields    []string // all other non-pk column field
}

// NewTable initializes a new table structure
func NewTable(tableName string, cs ...*Column) *Table {
	ts := &Table{
		Name:    tableName,
		Columns: Columns(cs),
	}
	return ts.update()
}

// update recalculates the internal cached fields
func (t *Table) update() *Table {
	if len(t.Columns) == 0 {
		return t
	}
	t.fieldsPK = t.Columns.PrimaryKeys().FieldNames()
	t.fieldsUNI = t.Columns.UniqueKeys().FieldNames()
	t.fields = t.Columns.ColumnsNoPK().FieldNames()
	t.CountPK = t.Columns.PrimaryKeys().Len()
	t.CountUnique = t.Columns.UniqueKeys().Len()
	return t
}

// LoadColumns reads the column information from the DB.
func (t *Table) LoadColumns(ctx context.Context, db Querier) (err error) {
	t.Columns, err = LoadColumns(ctx, db, t.Name)
	t.update()
	return errors.Wrapf(err, "[csdb] table.LoadColumns. Table %q", t.Name)
}

// TableAliasQuote returns a table name with the alias. catalog_product_entity
// with alias e would become `catalog_product_entity` AS `e`.
func (t *Table) TableAliasQuote(alias string) string {
	return dbr.Quoter.QuoteAs(t.Name, alias)
}

// ColumnAliasQuote prefixes non-id columns with an alias and puts quotes around
// them. Returns a copy.
func (t *Table) ColumnAliasQuote(alias string) []string {
	sl := make([]string, len(t.fields))
	copy(sl, t.fields)
	return dbr.Quoter.TableColumnAlias(alias, sl...)
}

// AllColumnAliasQuote prefixes all columns with an alias and puts quotes around
// them. Returns a copy.
func (t *Table) AllColumnAliasQuote(alias string) []string {
	sl := make([]string, len(t.fieldsPK)+len(t.fields))
	n := copy(sl, t.fieldsPK)
	copy(sl[n:], t.fields)
	return dbr.Quoter.TableColumnAlias(alias, sl...)
}

// In checks if column name n is a column of this table. Case sensitive.
func (t *Table) In(n string) bool {
	for _, c := range t.fieldsPK {
		if c == n {
			return true
		}
	}
	for _, c := range t.fields {
		if c == n {
			return true
		}
	}
	return false
}

// Select generates a SELECT * FROM tableName statement.
// DEPRECATED
func (t *Table) Select(dbrSess dbr.SessionRunner) (*dbr.SelectBuilder, error) {
	if t == nil {
		return nil, errors.NewFatalf("[csdb] Table cannot be nil")
	}
	return dbrSess.
		Select(t.AllColumnAliasQuote(MainTable)...).
		From(t.Name, MainTable), nil
}

// LoadSlice performs a SELECT * FROM `tableName` query and puts the results
// into the pointer slice `dest`. Returns the number of loaded rows and nil or 0
// and an error. The variadic thrid arguments can modify the SQL query.
// DEPRECATED
func (t *Table) LoadSlice(dbrSess dbr.SessionRunner, dest interface{}, cbs ...dbr.SelectCb) (int, error) {
	sb, err := t.Select(dbrSess)
	if err != nil {
		return 0, errors.Wrap(err, "[csdb] LoadSlice.Select")
	}

	for _, cb := range cbs {
		sb = cb(sb)
	}
	return sb.LoadStructs(dest)
}
