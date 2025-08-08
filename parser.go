package zconfig

import (
	"encoding"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseString(raw, res interface{}) (err error) {
	s, ok := raw.(string)
	if !ok {
		return ErrNotParseable
	}

	switch res := res.(type) {
	case *regexp.Regexp:
		var v *regexp.Regexp
		v, err = regexp.Compile(s)
		if err != nil {
			return
		}
		*res = *v
		return
	case encoding.TextUnmarshaler:
		return res.UnmarshalText([]byte(s))
	case encoding.BinaryUnmarshaler:
		return res.UnmarshalBinary([]byte(s))
	case *string:
		*res = s
		return nil
	case *[]byte:
		*res = []byte(s)
		return nil
	case *[]string:
		for _, c := range strings.Split(s, ",") {
			v := strings.TrimSpace(c)
			if v == "" {
				continue
			}
			*res = append(*res, v)
		}
		return nil
	case *[]int:
		for _, c := range strings.Split(s, ",") {
			raw := strings.TrimSpace(c)
			if raw == "" {
				continue
			}
			var v int
			v, err = strconv.Atoi(raw)
			if err != nil {
				return
			}
			*res = append(*res, v)
		}
	case *[]int64:
		for _, c := range strings.Split(s, ",") {
			raw := strings.TrimSpace(c)
			if raw == "" {
				continue
			}
			var v int64
			v, err = strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return
			}
			*res = append(*res, v)
		}
	case *bool:
		*res, err = strconv.ParseBool(s)
		return
	case *int:
		var v int64
		v, err = strconv.ParseInt(s, 10, strconv.IntSize)
		*res = int(v)
		return
	case *int8:
		var v int64
		v, err = strconv.ParseInt(s, 10, 8)
		*res = int8(v)
		return
	case *int16:
		var v int64
		v, err = strconv.ParseInt(s, 10, 16)
		*res = int16(v)
		return
	case *int32:
		var v int64
		v, err = strconv.ParseInt(s, 10, 32)
		*res = int32(v)
		return
	case *int64:
		var v int64
		v, err = strconv.ParseInt(s, 10, 64)
		*res = v
		return
	case *uint:
		var v uint64
		v, err = strconv.ParseUint(s, 10, strconv.IntSize)
		*res = uint(v)
		return
	case *uint8:
		var v uint64
		v, err = strconv.ParseUint(s, 10, 8)
		*res = uint8(v)
		return
	case *uint16:
		var v uint64
		v, err = strconv.ParseUint(s, 10, 16)
		*res = uint16(v)
		return
	case *uint32:
		var v uint64
		v, err = strconv.ParseUint(s, 10, 32)
		*res = uint32(v)
		return
	case *uint64:
		var v uint64
		v, err = strconv.ParseUint(s, 10, 64)
		*res = v
		return
	case *float32:
		var v float64
		v, err = strconv.ParseFloat(s, 32)
		*res = float32(v)
		return
	case *float64:
		var v float64
		v, err = strconv.ParseFloat(s, 64)
		*res = v
		return
	case *time.Duration:
		var v time.Duration
		v, err = time.ParseDuration(s)
		if err != nil {
			return
		}
		*res = v
	default:
		return ErrNotParseable
	}

	return nil
}
