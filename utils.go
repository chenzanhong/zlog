package zlog

import (
	"time"

	"go.uber.org/zap"
)

// toZapFields 将 zlog.Field 转为 zap.Field
func toZapFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zfs := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch f.typ {
		case fieldString:
			zfs = append(zfs, zap.String(f.key, f.value.(string)))
		case fieldInt:
			zfs = append(zfs, zap.Int(f.key, f.value.(int)))
		case fieldInt64:
			zfs = append(zfs, zap.Int64(f.key, f.value.(int64)))
		case fieldBool:
			zfs = append(zfs, zap.Bool(f.key, f.value.(bool)))
		case fieldFloat64:
			zfs = append(zfs, zap.Float64(f.key, f.value.(float64)))
		case fieldDuration:
			zfs = append(zfs, zap.Duration(f.key, f.value.(time.Duration)))
		case fieldTime:
			zfs = append(zfs, zap.Time(f.key, f.value.(time.Time)))
		default:
			zfs = append(zfs, zap.Any(f.key, f.value))
		}
	}
	return zfs
}
