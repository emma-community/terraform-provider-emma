package tools

import (
	"encoding/json"
	"io"
	"net/http"
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
