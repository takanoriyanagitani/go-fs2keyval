package fs2kv

import (
	"context"
	"path"
	"regexp"
	"unicode/utf8"

	s2k "github.com/takanoriyanagitani/go-sql2keyval"
)

type KeyValidator func(key []byte) bool

type PairValidator func(p s2k.Pair) bool

type BatchValidator func(b s2k.Batch) bool

type FileLike struct {
	Path string // bucket name + "/" + key(utf8 byte string)
	Val  []byte
}

type Batch2FileLike func(b s2k.Batch) s2k.Option[FileLike]

type SetFilelikeBatch func(ctx context.Context, many s2k.Iter[FileLike]) error

var Utf8validator KeyValidator = utf8.Valid

func RegexpValidatorNewMust(pat string) KeyValidator {
	re := regexp.MustCompile(pat)
	return re.Match
}

func MultiValidatorNew(v []KeyValidator) KeyValidator {
	return func(key []byte) bool {
		var i s2k.Iter[KeyValidator] = s2k.IterFromArray(v)
		return s2k.IterReduce(i, true, func(state bool, item KeyValidator) bool {
			invalid := !state
			if invalid {
				return false
			}
			return item(key)
		})
	}
}

func PairValidatorFromKV(kv KeyValidator) PairValidator {
	return func(p s2k.Pair) bool {
		return kv(p.Key)
	}
}

func BatchValidatorFromPV(pv PairValidator) BatchValidator {
	return func(b s2k.Batch) bool {
		return pv(b.Pair())
	}
}

type Bytes2string func(b []byte) s2k.Option[string]

func Batch2FilelikeNew(b2s Bytes2string) Batch2FileLike {
	return func(b s2k.Batch) s2k.Option[FileLike] {
		var p s2k.Pair = b.Pair()
		var k []byte = p.Key
		var ko s2k.Option[string] = b2s(k)
		return s2k.OptionMap(ko, func(ks string) FileLike {
			return FileLike{
				Path: path.Join(b.Bucket(), ks),
				Val:  p.Val,
			}
		})
	}
}

func ErrorWarn(funcError func() error, funcWarn func() error) error {
	e := funcError()
	if nil == e {
		return funcWarn()
	}
	defer funcWarn()
	return e
}

func Error1st(f []func() error) error {
	return s2k.IterReduce(s2k.IterFromArray(f), nil, func(e error, item func() error) error {
		if nil == e {
			return item()
		}
		return e
	})
}

func IfOk(e error, f func() error) error {
	return Error1st([]func() error{
		func() error { return e },
		f,
	})
}
