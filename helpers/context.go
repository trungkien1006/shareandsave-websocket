package helpers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUintFromContext(c *gin.Context, key string) (uint, error) {
	val, exists := c.Get(key)
	if !exists {
		return 0, errors.New("key không tồn tại trong context")
	}

	switch v := val.(type) {
	case uint:
		return v, nil
	case int:
		return uint(v), nil
	case float64:
		return uint(v), nil
	case string:
		parsedID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, errors.New("chuỗi không hợp lệ")
		}
		return uint(parsedID), nil
	default:
		return 0, errors.New("kiểu dữ liệu không hỗ trợ")
	}
}

func GetStringFromContext(c *gin.Context, key string) (string, error) {
	val, exists := c.Get(key)
	if !exists {
		return "", fmt.Errorf("key '%s' không tồn tại trong context", key)
	}

	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", errors.New("giá trị không phải kiểu string hoặc không thể chuyển đổi")
	}
}
