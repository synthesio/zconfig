package zconfig

import (
	"encoding"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CustomParser interface {
	Parse(string) error
}

func ParseString(raw, res interface{}) (err error) {
	s, ok := raw.(string)
	if !ok {
		return ErrNotParseable
	}

	switch res := res.(type) {
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
			v, err := strconv.Atoi(raw)
			if err != nil {
				return err
			}
			*res = append(*res, v)
		}
	case *[]int64:
		for _, c := range strings.Split(s, ",") {
			raw := strings.TrimSpace(c)
			if raw == "" {
				continue
			}
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return err
			}
			*res = append(*res, v)
		}
	case *bool:
		*res, err = strconv.ParseBool(s)
		return err
	case *int:
		v, err := strconv.ParseInt(s, 10, strconv.IntSize)
		*res = int(v)
		return err
	case *int8:
		v, err := strconv.ParseInt(s, 10, 8)
		*res = int8(v)
		return err
	case *int16:
		v, err := strconv.ParseInt(s, 10, 16)
		*res = int16(v)
		return err
	case *int32:
		v, err := strconv.ParseInt(s, 10, 32)
		*res = int32(v)
		return err
	case *int64:
		v, err := strconv.ParseInt(s, 10, 64)
		*res = int64(v)
		return err
	case *uint:
		v, err := strconv.ParseUint(s, 10, strconv.IntSize)
		*res = uint(v)
		return err
	case *uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		*res = uint8(v)
		return err
	case *uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		*res = uint16(v)
		return err
	case *uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		*res = uint32(v)
		return err
	case *uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		*res = uint64(v)
		return err
	case *float32:
		v, err := strconv.ParseFloat(s, 32)
		*res = float32(v)
		return err
	case *float64:
		v, err := strconv.ParseFloat(s, 64)
		*res = float64(v)
		return err
	case *regexp.Regexp:
		v, err := regexp.Compile(s)
		if err != nil {
			return err
		}
		*res = *v
		return nil
	case *time.Duration:
		v, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		*res = v
	default:
		if v, ok := res.(CustomParser); ok {
			err := v.Parse(s)
			if err == nil {
				return nil
			}
		}
		return ErrNotParseable
	}

	return nil
}
