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

package log

// Most of this code can be improved to fits the needs for others.
// Main reason for implementing this was to provide a basic leveled logger without
// any dependencies to third party packages.

import (
	"bytes"
	"fmt"
	"io"
	std "log"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
)

const (
	StdLevelFatal int = iota + 1
	StdLevelInfo
	StdLevelDebug
)

// StdLog implements logging with Go's standard library
type StdLog struct {
	gw      io.Writer // global writer
	level   int
	flag    int // global flag http://golang.org/pkg/log/#pkg-constants
	noTrace bool
	debug   *std.Logger
	info    *std.Logger
	fatal   *std.Logger
}

// StdOption can be used as an argument in NewStdLog to configure a standard logger.
type StdOption func(*StdLog)

// NewStdLog creates a new logger with 6 different sub loggers.
// You can use option functions to modify each logger independently.
// Default output goes to Stderr.
func NewStdLog(opts ...StdOption) *StdLog {
	sl := &StdLog{
		level: StdLevelInfo,
		gw:    os.Stderr,
		flag:  std.LstdFlags,
	}
	for _, o := range opts {
		o(sl)
	}
	if sl.debug == nil {
		sl.debug = std.New(sl.gw, "DEBUG ", sl.flag)
	}
	if sl.info == nil {
		sl.info = std.New(sl.gw, "INFO ", sl.flag)
	}
	if sl.fatal == nil {
		sl.fatal = std.New(sl.gw, "FATAL ", sl.flag)
	}
	return sl
}

// SetStdWriter sets the global writer for all loggers. This global writer can be
// overwritten by individual level options.
func WithStdWriter(w io.Writer) StdOption {
	return func(l *StdLog) {
		l.gw = w
	}
}

// WithStdFlag sets the global flag for all loggers.
// These flags define which text to prefix to each log entry generated by the Logger.
// This global flag can be overwritten by individual level options.
// Please see http://golang.org/pkg/log/#pkg-constants
func WithStdFlag(f int) StdOption {
	return func(l *StdLog) {
		l.flag = f
	}
}

// WithStdLevel sets the log level. See constants Level*
func WithStdLevel(level int) StdOption {
	return func(l *StdLog) {
		l.SetLevel(level)
	}
}

// WithStdDebug applies options for debug logging
func WithStdDebug(out io.Writer, prefix string, flag int) StdOption {
	return func(l *StdLog) {
		l.debug = std.New(out, prefix, flag)
	}
}

// WithStdInfo applies options for info logging
func WithStdInfo(out io.Writer, prefix string, flag int) StdOption {
	return func(l *StdLog) {
		l.info = std.New(out, prefix, flag)
	}
}

// WithStdFatal applies options for fatal logging
func WithStdFatal(out io.Writer, prefix string, flag int) StdOption {
	return func(l *StdLog) {
		l.fatal = std.New(out, prefix, flag)
	}
}

// SetStdDisableStackTrace disables the stack trace when an error will be passed
// as an argument.
func SetStdDisableStackTrace() StdOption {
	return func(l *StdLog) {
		l.noTrace = true
	}
}

// New returns a new Logger that has this logger's context plus the given context
// This function panics if an argument is not of type StdOption.
func (l *StdLog) New(iOpts ...interface{}) Logger {
	var opts = make([]StdOption, len(iOpts), len(iOpts))
	for i, iopt := range iOpts {
		if o, ok := iopt.(StdOption); ok {
			opts[i] = o
		} else {
			panic("Arguments to New() can only be StdOption types!")
		}
	}
	return NewStdLog(opts...)
}

// Debug logs a debug entry.
func (l *StdLog) Debug(msg string, args ...interface{}) {
	l.log(StdLevelDebug, msg, args)
}

// Info logs an info entry.
func (l *StdLog) Info(msg string, args ...interface{}) {
	l.log(StdLevelInfo, msg, args)
}

// Fatal logs a fatal entry then panics.
func (l *StdLog) Fatal(msg string, args ...interface{}) {
	l.log(StdLevelFatal, msg, args)
}

// log logs a leveled entry. Panics if an unknown level has been provided.
func (l *StdLog) log(level int, msg string, args []interface{}) {
	if l.level >= level {
		switch level {
		case StdLevelDebug:
			// l.debug.Print(stdFormat(msg, append(args, "in", getStackTrace())))
			l.debug.Print(stdFormat(msg, args, l.noTrace))
			break
		case StdLevelInfo:
			l.info.Print(stdFormat(msg, args, l.noTrace))
			break
		case StdLevelFatal:
			l.fatal.Panic(stdFormat(msg, args, l.noTrace))
			break
		default:
			panic("Unknown Log Level")
		}
	}
}

// IsDebug determines if this logger logs a debug statement.
func (l *StdLog) IsDebug() bool { return l.level >= StdLevelDebug }

// IsInfo determines if this logger logs an info statement.
func (l *StdLog) IsInfo() bool { return l.level >= StdLevelInfo }

// SetLevel sets the level of this logger.
func (l *StdLog) SetLevel(level int) {
	l.level = level
}

func getStackTrace() string {
	s := debug.Stack()
	lb := []byte("\n")
	parts := bytes.Split(s, lb)
	return string(bytes.Join(parts[6:], lb))
}

// Following Code by: https://github.com/mgutz Mario Gutierrez / MIT License
// And some changes by @SchumacherFM

var pool = newBP()

// AssignmentChar represents the assignment character between key-value pairs
var AssignmentChar = ": "

// Separator is the separator to use between key value pairs
var Separator = " "

func stdSetKV(buf *bytes.Buffer, key string, val interface{}, noTrace bool) {
	buf.WriteString(Separator)
	buf.WriteString(key)
	buf.WriteString(AssignmentChar)
	if err, ok := val.(error); ok {
		buf.WriteString(err.Error())
		if false == noTrace {
			buf.WriteRune('\n')
			buf.WriteString(getStackTrace())
		}
		return
	}
	buf.WriteString(fmt.Sprintf("%#v", val))
}

func stdFormat(msg string, args []interface{}, noTrace bool) string {
	buf := pool.Get()
	defer pool.Put(buf)

	buf.WriteString(msg)
	lenArgs := len(args)
	if lenArgs > 0 {
		if lenArgs == 1 {
			stdSetKV(buf, "_", args[0], noTrace)
		} else if lenArgs%2 == 0 {
			for i := 0; i < lenArgs; i += 2 {
				if key, ok := args[i].(string); ok {
					if key == "" {
						// show key is invalid
						stdSetKV(buf, badKeyAtIndex(i), args[i+1], noTrace)
					} else {
						stdSetKV(buf, key, args[i+1], noTrace)
					}
				} else {
					// show key is invalid
					stdSetKV(buf, badKeyAtIndex(i), args[i+1], noTrace)
				}
			}
		} else {
			stdSetKV(buf, `FIX_IMBALANCED_PAIRS`, args, noTrace)
		}
	}
	buf.WriteRune('\n')
	return buf.String()
}

func badKeyAtIndex(i int) string {
	return "BAD_KEY_AT_INDEX_" + strconv.Itoa(i)
}

type bp struct {
	sync.Pool
}

func newBP() *bp {
	return &bp{
		Pool: sync.Pool{New: func() interface{} {
			b := bytes.NewBuffer(make([]byte, 128))
			b.Reset()
			return b
		}},
	}
}

func (bp *bp) Get() *bytes.Buffer {
	return bp.Pool.Get().(*bytes.Buffer)
}

func (bp *bp) Put(b *bytes.Buffer) {
	b.Reset()
	bp.Pool.Put(b)
}
