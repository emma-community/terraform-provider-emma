package tools

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExtractErrorMessage(t *testing.T) {
	response := "{\"message\":\"testMessage\", \"code\":\"test\"}"
	var httpResponse http.Response
	reader := strings.NewReader(response)
	readCloser := io.NopCloser(reader)
	httpResponse.Body = readCloser
	extractedMessage := ExtractErrorMessage(&httpResponse)
	assert.Equal(t, "testMessage", extractedMessage)
	readCloser.Close()
}

func TestStringToInt32(t *testing.T) {
	str := "42"
	assert.Equal(t, int32(42), StringToInt32(str))
}
