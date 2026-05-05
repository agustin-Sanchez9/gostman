package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Response holds the result of an HTTP request.
type Response struct {
	StatusCode int
	Status     string
	Headers    string
	Body       string
}

// SendRequest performs an HTTP request with the given parameters.
// Headers are provided in bulk edit format: one "Key: Value" per line.
func SendRequest(method, url, headers, body string) (*Response, error) {
	// Parse bulk headers
	headerMap := make(http.Header)
	lines := strings.Split(headers, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headerMap.Add(key, value)
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	for key, values := range headerMap {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Pretty-print JSON response bodies when possible
	respBodyStr := string(respBody)
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, respBody, "", "  "); err == nil {
			respBodyStr = prettyJSON.String()
		}
	}

	// Format response headers
	var headerStr strings.Builder
	for key, values := range resp.Header {
		for _, value := range values {
			headerStr.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headerStr.String(),
		Body:       respBodyStr,
	}, nil
}
