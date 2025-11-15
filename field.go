package zlog

import (
	"time"
)

type fieldType int

const (
	fieldAny fieldType = iota
	fieldString
	fieldInt
	fieldInt64
	fieldBool
	fieldFloat64
	fieldDuration
	fieldTime
)

// Field is zlog's custom log field type, hiding zap.Field internally
type Field struct {
	key   string
	value interface{}
	typ   fieldType
}

// Constructor functions
func String(key, val string) Field { return Field{key: key, value: val, typ: fieldString} }
func Int(key string, val int) Field { return Field{key: key, value: val, typ: fieldInt} }
func Int64(key string, val int64) Field { return Field{key: key, value: val, typ: fieldInt64} }
func Bool(key string, val bool) Field { return Field{key: key, value: val, typ: fieldBool} }
func Float64(key string, val float64) Field { return Field{key: key, value: val, typ: fieldFloat64} }
func Duration(key string, val time.Duration) Field { return Field{key: key, value: val, typ: fieldDuration} }
func Time(key string, val time.Time) Field { return Field{key: key, value: val, typ: fieldTime} }
func Any(key string, val interface{}) Field { return Field{key: key, value: val, typ: fieldAny} }