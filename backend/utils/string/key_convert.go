package strutil

import (
	"strconv"
)

// AnyToInt convert any value to int
func AnyToInt(val any) (int, bool) {
	switch val.(type) {
	case string:
		num, err := strconv.Atoi(val.(string))
		if err != nil {
			return 0, false
		}
		return num, true
	case float64:
		return int(val.(float64)), true
	case float32:
		return int(val.(float32)), true
	case int64:
		return int(val.(int64)), true
	case int32:
		return int(val.(int32)), true
	case int:
		return val.(int), true
	case bool:
		if val.(bool) {
			return 1, true
		} else {
			return 0, true
		}
	}
	return 0, false
}
