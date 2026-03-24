package filler

import (
	"strings"
	"time"
)

// TagToFloat converts various types to float64
func TagToFloat(val any) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case []float64:
		if len(v) > 0 {
			return v[0]
		}
	case []uint32:
		if len(v) > 0 {
			return float64(v[0])
		}
	}
	return 0
}

// TagToInt converts various types to int
func TagToInt(val any) int {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case int16:
		return int(v)
	case int32:
		return int(v)
	case uint16:
		return int(v)
	case uint32:
		return int(v)
	case []uint16:
		if len(v) > 0 {
			return int(v[0])
		}
	}
	return 0
}

// TagToTime tries to parse a time value or string representation
func TagToTime(val any, str string) *time.Time {
	if t, ok := val.(time.Time); ok {
		return &t
	}
	str = strings.Trim(str, "\x00 ")
	if str == "" {
		return nil
	}

	formats := []string{
		"2006:01:02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return &t
		}
	}
	return nil
}
