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

package store_test

import (
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/util/null"
)

func init() {
	store.TableCollection = csdb.MustNewTables(
		csdb.WithTable(
			store.TableIndexStore,
			"store",
			&csdb.Column{Field: (`store_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`PRI`), Extra: (`auto_increment`)},
			&csdb.Column{Field: (`code`), ColumnType: (`varchar(32)`), Null: (`YES`), Key: (`UNI`), Extra: (``)},
			&csdb.Column{Field: (`website_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`group_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`name`), ColumnType: (`varchar(255)`), Null: (`NO`), Key: (``), Extra: (``)},
			&csdb.Column{Field: (`sort_order`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (``), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`is_active`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
		),
		csdb.WithTable(
			store.TableIndexGroup,
			"store_group",
			&csdb.Column{Field: (`group_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`PRI`), Extra: (`auto_increment`)},
			&csdb.Column{Field: (`website_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`name`), ColumnType: (`varchar(255)`), Null: (`NO`), Key: (``), Extra: (``)},
			&csdb.Column{Field: (`root_category_id`), ColumnType: (`int(10) unsigned`), Null: (`NO`), Key: (``), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`default_store_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
		),
		csdb.WithTable(
			store.TableIndexWebsite,
			"store_website",
			&csdb.Column{Field: (`website_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`PRI`), Extra: (`auto_increment`)},
			&csdb.Column{Field: (`code`), ColumnType: (`varchar(32)`), Null: (`YES`), Key: (`UNI`), Extra: (``)},
			&csdb.Column{Field: (`name`), ColumnType: (`varchar(64)`), Null: (`YES`), Key: (``), Extra: (``)},
			&csdb.Column{Field: (`sort_order`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`default_group_id`), ColumnType: (`smallint(5) unsigned`), Null: (`NO`), Key: (`MUL`), Default: null.StringFrom(`0`), Extra: (``)},
			&csdb.Column{Field: (`is_default`), ColumnType: (`smallint(5) unsigned`), Null: (`YES`), Key: (``), Default: null.StringFrom(`0`), Extra: (``)},
		),
	)
}
