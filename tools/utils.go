package tools

import (
	"fmt"
	"strconv"
)

func ConvertToString(val interface{}) string {
	if val == nil {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return str
}

func ConvertToInt64(val interface{}) int64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

func ConvertToFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return float64(v)
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return float64(i)
	default:
		return 0
	}
}

func ConvertToBool(val interface{}) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	default:
		return false
	}
}
