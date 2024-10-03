package tools

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"io"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ExtractErrorMessage(response *http.Response) string {
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return ""
	}
	data := ErrorResponse{}
	err = json.Unmarshal(responseBytes, &data)
	if err != nil {
		return ""
	}
	return data.Message
}

func StringToInt32(value string) int32 {
	num, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		panic(err)
	}
	return int32(num)
}

func Int64ToInt32Pointer(value int64) *int32 {
	val := int32(value)
	return &val
}

func Float64ToFloat32Pointer(value float64) *float32 {
	val := float32(value)
	return &val
}

func ToPointer[T any](value T) *T {
	return &value
}

func ToInt32PointerOrNil(configValue types.Int64) *int32 {
	if !configValue.IsUnknown() && !configValue.IsNull() {
		return Int64ToInt32Pointer(configValue.ValueInt64())
	}
	return nil
}

func ToFloat32PointerOrNil(configValue types.Float64) *float32 {
	if !configValue.IsUnknown() && !configValue.IsNull() {
		return Float64ToFloat32Pointer(configValue.ValueFloat64())
	}
	return nil
}

func GetFloat64OrDefault(configValue *float32, defaultValue types.Float64) types.Float64 {
	if configValue != nil {
		return types.Float64Value(float64(*configValue))
	}
	return defaultValue
}

func GetInt64OrDefault(configValue *int32, defaultValue types.Int64) types.Int64 {
	if configValue != nil {
		return types.Int64Value(int64(*configValue))
	}
	return defaultValue
}
