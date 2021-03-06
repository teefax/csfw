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

package zapw_test

import (
	"bytes"
	"math"
	"testing"
	"time"

	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/log/zapw"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/zap"
)

var _ log.Logger = (*zapw.Wrap)(nil)

func getZap(lvl zap.Level) (*bytes.Buffer, log.Logger) {
	buf := &bytes.Buffer{}
	l := zapw.Wrap{
		Zap: zap.New(zap.NewJSONEncoder(zap.NoTime()), lvl, zap.Output(zap.AddSync(buf)), zap.ErrorOutput(zap.AddSync(buf)), zap.Fields(zap.Int("answer", 42))),
	}

	return buf, l
}

func getZapWithLog(lvl zap.Level) string {
	buf, l := getZap(lvl)

	if l.IsDebug() {
		l.Debug("log_15_debug", log.Err(errors.New("I'm an debug error")), log.Float64("pi", 3.14159), log.String("kDebug", "v1"), log.Duration("debugDur", time.Minute))
	}
	if l.IsInfo() {
		l.Info("log_15_info", log.Err(errors.New("I'm an info error")), log.Float64("e", 2.7182), log.String("kInfo", "v1"), log.Duration("infoDur", time.Hour))
	}
	return buf.String()
}

func TestNewJSON_Debug(t *testing.T) {
	out := getZapWithLog(zap.DebugLevel)
	assert.Contains(t, out, `"answer":42`)
	assert.Contains(t, out, `"error":"I'm an debug error"`)
	assert.Contains(t, out, `"pi":3.14159`)
	assert.Contains(t, out, `"debugDur":600000000`)
	assert.Contains(t, out, `"pi":3.14159`)
	assert.Contains(t, out, `"error":"I'm an info error"`)
	assert.Contains(t, out, `"kInfo":"v1"`)
	assert.Contains(t, out, `"infoDur":3600000000`)
}

func TestNewJSON_Info(t *testing.T) {
	out := getZapWithLog(zap.InfoLevel)
	assert.NotContains(t, out, `ds":{"answer":42,"error":"I'm an debug error","pi":3.14159,"kDebug":"v1","debugDur":600000000`)
	assert.Contains(t, out, `"error":"I'm an info error"`)
	assert.Contains(t, out, `"kInfo":"v1","infoDur":3600000000`)
	assert.Contains(t, out, `"e":2.7182`)
}

type marshalMock struct {
	string
	float64
	bool
	error
}

func (mm marshalMock) MarshalLog(kv log.KeyValuer) error {
	kv.AddBool("kvbool", mm.bool)
	kv.AddString("kvstring", mm.string)
	kv.AddFloat64("kvfloat64", mm.float64)
	return mm.error
}

func TestAddMarshaler(t *testing.T) {
	buf, l := getZap(zap.DebugLevel)

	l.Debug("log_15_debug", log.Err(errors.New("I'm an debug error")), log.Float64("pi", 3.14159))

	l.Debug("log_15_marshalling", log.Object("anObject", 42), log.Marshal("marshalLogMock", marshalMock{
		string:  "s1",
		float64: math.Ln2,
		bool:    true,
	}))
	assert.Contains(t, buf.String(), `"anObject":42`)
	assert.Contains(t, buf.String(), `"kvfloat64":0.6931471805599453`)
	assert.Contains(t, buf.String(), `"kvstring":"s1"`)
}

func TestAddMarshaler_Error(t *testing.T) {
	buf, l := getZap(zap.DebugLevel)

	l.Debug("marshalling", log.Marshal("marshalLogMock", marshalMock{
		error: errors.New("Whooops"),
	}))
	assert.Contains(t, buf.String(), `"level":"debug","msg":"marshalling","answer":42,"kvbool":false,"kvstring":"","kvfloat64":0,"error":"Whooops\ngithub.com/corestoreio/csfw/log/zapw_test.TestAddMarshaler_Error`)
}
