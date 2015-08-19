// Copyright 2015, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package store

import (
	"io"

	"github.com/ugorji/go/codec"
)

// @todo add FromJSON

var codecJSON codec.Handle = new(codec.JsonHandle)

// ToJSON fast JSON encoding with http://ugorji.net/blog/go-codec-primer algorithm.
func (s *Store) ToJSON(w io.Writer) error {
	var enc *codec.Encoder = codec.NewEncoder(w, codecJSON)
	return enc.Encode(s.Data)
}

// ToJSON fast JSON encoding with http://ugorji.net/blog/go-codec-primer algorithm.
func (ws *Website) ToJSON(w io.Writer) error {
	var enc *codec.Encoder = codec.NewEncoder(w, codecJSON)
	return enc.Encode(ws.Data)
}

// ToJSON fast JSON encoding with http://ugorji.net/blog/go-codec-primer algorithm.
func (g *Group) ToJSON(w io.Writer) error {
	var enc *codec.Encoder = codec.NewEncoder(w, codecJSON)
	return enc.Encode(g.Data)
}

// @todo add other encoding/decoding algorithms