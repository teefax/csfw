// DO NOT EDIT!
// Code generated by ffjson <https://github.com/pquerna/ffjson>
// source: claim.go
// DO NOT EDIT!

package jwtclaim

import (
	"bytes"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

func (mj *Store) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if mj == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := mj.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (mj *Store) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if mj == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{ `)
	if len(mj.Store) != 0 {
		buf.WriteString(`"store":`)
		fflib.WriteJsonString(buf, string(mj.Store))
		buf.WriteByte(',')
	}
	if len(mj.UserID) != 0 {
		buf.WriteString(`"userid":`)
		fflib.WriteJsonString(buf, string(mj.UserID))
		buf.WriteByte(',')
	}
	if len(mj.Audience) != 0 {
		buf.WriteString(`"aud":`)
		fflib.WriteJsonString(buf, string(mj.Audience))
		buf.WriteByte(',')
	}
	if mj.ExpiresAt != 0 {
		buf.WriteString(`"exp":`)
		fflib.FormatBits2(buf, uint64(mj.ExpiresAt), 10, mj.ExpiresAt < 0)
		buf.WriteByte(',')
	}
	if len(mj.ID) != 0 {
		buf.WriteString(`"jti":`)
		fflib.WriteJsonString(buf, string(mj.ID))
		buf.WriteByte(',')
	}
	if mj.IssuedAt != 0 {
		buf.WriteString(`"iat":`)
		fflib.FormatBits2(buf, uint64(mj.IssuedAt), 10, mj.IssuedAt < 0)
		buf.WriteByte(',')
	}
	if len(mj.Issuer) != 0 {
		buf.WriteString(`"iss":`)
		fflib.WriteJsonString(buf, string(mj.Issuer))
		buf.WriteByte(',')
	}
	if mj.NotBefore != 0 {
		buf.WriteString(`"nbf":`)
		fflib.FormatBits2(buf, uint64(mj.NotBefore), 10, mj.NotBefore < 0)
		buf.WriteByte(',')
	}
	if len(mj.Subject) != 0 {
		buf.WriteString(`"sub":`)
		fflib.WriteJsonString(buf, string(mj.Subject))
		buf.WriteByte(',')
	}
	buf.Rewind(1)
	buf.WriteByte('}')
	return nil
}

const (
	ffj_t_Storebase = iota
	ffj_t_Storeno_such_key

	ffj_t_Store_Store

	ffj_t_Store_UserID

	ffj_t_Store_Audience

	ffj_t_Store_ExpiresAt

	ffj_t_Store_ID

	ffj_t_Store_IssuedAt

	ffj_t_Store_Issuer

	ffj_t_Store_NotBefore

	ffj_t_Store_Subject
)

var ffj_key_Store_Store = []byte("store")

var ffj_key_Store_UserID = []byte("userid")

var ffj_key_Store_Audience = []byte("aud")

var ffj_key_Store_ExpiresAt = []byte("exp")

var ffj_key_Store_ID = []byte("jti")

var ffj_key_Store_IssuedAt = []byte("iat")

var ffj_key_Store_Issuer = []byte("iss")

var ffj_key_Store_NotBefore = []byte("nbf")

var ffj_key_Store_Subject = []byte("sub")

func (uj *Store) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return uj.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

func (uj *Store) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error = nil
	currentKey := ffj_t_Storebase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffj_t_Storeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'a':

					if bytes.Equal(ffj_key_Store_Audience, kn) {
						currentKey = ffj_t_Store_Audience
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'e':

					if bytes.Equal(ffj_key_Store_ExpiresAt, kn) {
						currentKey = ffj_t_Store_ExpiresAt
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'i':

					if bytes.Equal(ffj_key_Store_IssuedAt, kn) {
						currentKey = ffj_t_Store_IssuedAt
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffj_key_Store_Issuer, kn) {
						currentKey = ffj_t_Store_Issuer
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'j':

					if bytes.Equal(ffj_key_Store_ID, kn) {
						currentKey = ffj_t_Store_ID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'n':

					if bytes.Equal(ffj_key_Store_NotBefore, kn) {
						currentKey = ffj_t_Store_NotBefore
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 's':

					if bytes.Equal(ffj_key_Store_Store, kn) {
						currentKey = ffj_t_Store_Store
						state = fflib.FFParse_want_colon
						goto mainparse

					} else if bytes.Equal(ffj_key_Store_Subject, kn) {
						currentKey = ffj_t_Store_Subject
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'u':

					if bytes.Equal(ffj_key_Store_UserID, kn) {
						currentKey = ffj_t_Store_UserID
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.EqualFoldRight(ffj_key_Store_Subject, kn) {
					currentKey = ffj_t_Store_Subject
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Store_NotBefore, kn) {
					currentKey = ffj_t_Store_NotBefore
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Store_Issuer, kn) {
					currentKey = ffj_t_Store_Issuer
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Store_IssuedAt, kn) {
					currentKey = ffj_t_Store_IssuedAt
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Store_ID, kn) {
					currentKey = ffj_t_Store_ID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Store_ExpiresAt, kn) {
					currentKey = ffj_t_Store_ExpiresAt
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffj_key_Store_Audience, kn) {
					currentKey = ffj_t_Store_Audience
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Store_UserID, kn) {
					currentKey = ffj_t_Store_UserID
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.EqualFoldRight(ffj_key_Store_Store, kn) {
					currentKey = ffj_t_Store_Store
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffj_t_Storeno_such_key
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffj_t_Store_Store:
					goto handle_Store

				case ffj_t_Store_UserID:
					goto handle_UserID

				case ffj_t_Store_Audience:
					goto handle_Audience

				case ffj_t_Store_ExpiresAt:
					goto handle_ExpiresAt

				case ffj_t_Store_ID:
					goto handle_ID

				case ffj_t_Store_IssuedAt:
					goto handle_IssuedAt

				case ffj_t_Store_Issuer:
					goto handle_Issuer

				case ffj_t_Store_NotBefore:
					goto handle_NotBefore

				case ffj_t_Store_Subject:
					goto handle_Subject

				case ffj_t_Storeno_such_key:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Store:

	/* handler: uj.Store type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Store = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_UserID:

	/* handler: uj.UserID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.UserID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Audience:

	/* handler: uj.Audience type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Audience = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ExpiresAt:

	/* handler: uj.ExpiresAt type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.ExpiresAt = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_ID:

	/* handler: uj.ID type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.ID = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_IssuedAt:

	/* handler: uj.IssuedAt type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.IssuedAt = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Issuer:

	/* handler: uj.Issuer type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Issuer = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_NotBefore:

	/* handler: uj.NotBefore type=int64 kind=int64 quoted=false*/

	{
		if tok != fflib.FFTok_integer && tok != fflib.FFTok_null {
			return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for int64", tok))
		}
	}

	{

		if tok == fflib.FFTok_null {

		} else {

			tval, err := fflib.ParseInt(fs.Output.Bytes(), 10, 64)

			if err != nil {
				return fs.WrapErr(err)
			}

			uj.NotBefore = int64(tval)

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Subject:

	/* handler: uj.Subject type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			uj.Subject = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:
	return nil
}