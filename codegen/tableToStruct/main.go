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

package main

import (
	"sync"

	"github.com/corestoreio/csfw/codegen"
	"github.com/corestoreio/csfw/storage/csdb"
	"github.com/corestoreio/csfw/storage/dbr"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/corestoreio/csfw/util/log"
)

// PkgLog global package based logger
var PkgLog log.Logger = log.NewStdLog(
	log.WithStdLevel(log.StdLevelFatal),
	//log.SetStdLevel(log.StdLevelDebug),
)

func init() {
	codegen.PkgLog = PkgLog
	log.PkgLog = PkgLog
}

// Connect creates a new database connection from a DSN stored in an
// environment variable CS_DSN.
func Connect(opts ...dbr.ConnectionOption) (*dbr.Connection, error) {
	c, err := dbr.NewConnection(dbr.WithDSN(csdb.MustGetDSN()))
	if err != nil {
		return nil, errors.Wrap(err, "[csdb] dbr.NewConnection")
	}
	if err := c.Options(opts...); err != nil {
		return nil, errors.Wrap(err, "[csdb] dbr.NewConnection.Options")
	}
	return c, err
}

func main() {
	defer log.WhenDone(PkgLog).Info("Stats")
	dbc, err := Connect()
	codegen.LogFatal(err)
	defer dbc.Close()
	var wg sync.WaitGroup

	mageVersion := detectMagentoVersion(dbc.NewSession())

	for _, tStruct := range codegen.ConfigTableToStruct {
		go newGenerator(tStruct, dbc, &wg).setMagentoVersion(mageVersion).run()
	}

	wg.Wait()

	//	for _, ts := range codegen.ConfigTableToStruct {
	//		// due to a race condition the codec generator must run after the newGenerator() calls
	//		// TODO(cs) fix https://github.com/ugorji/go/issues/92#issuecomment-140410732
	//		runCodec(ts.Package, ts.OutputFile.AppendName("_codec").String(), ts.OutputFile.String())
	//	}
}
