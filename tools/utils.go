package tools

import (
	"encoding/json"
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
