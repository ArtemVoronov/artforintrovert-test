//go:build integration
// +build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/ArtemVoronov/artforintrovert-test/internal/services/records"
)

type TestHttpClient struct {
}

var testHttpClient TestHttpClient = TestHttpClient{}

func (p *TestHttpClient) UpsertRecord(id any, data any) (int, string, error) {
	body, err := CreateRecordBody(id, data)
	if err != nil {
		return -1, "", err
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/records/", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	TestRouter.ServeHTTP(w, req)
	return w.Code, w.Body.String(), nil
}

func (p *TestHttpClient) DeleteRecord(id any) (int, string, error) {
	body, err := CreateRecordBody(id, nil)
	if err != nil {
		return -1, "", err
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/records/", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	TestRouter.ServeHTTP(w, req)
	return w.Code, w.Body.String(), nil
}

func (p *TestHttpClient) GetAllRecords() (int, string, error) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/records/", nil)
	req.Header.Set("Content-Type", "application/json")
	TestRouter.ServeHTTP(w, req)
	return w.Code, w.Body.String(), nil
}

func ParseForJsonBody(paramName string, paramValue any) (string, error) {
	result := ""
	switch paramType := paramValue.(type) {
	case int:
		result = "\"" + paramName + "\": " + strconv.Itoa(paramValue.(int))
	case string:
		result = "\"" + paramName + "\": \"" + paramValue.(string) + "\""
	case nil:
		result = ""
	default:
		return "", fmt.Errorf("unkown type for '%s': %v", paramName, paramType)
	}
	return result, nil
}

func CreateRecordBody(id any, data any) (string, error) {
	idField, err := ParseForJsonBody("id", id)
	if err != nil {
		return "", err
	}
	dataField, err := ParseForJsonBody("data", data)
	if err != nil {
		return "", err
	}

	result := "{"
	if idField != "" {
		result += idField + ","
	}
	if dataField != "" {
		result += dataField + ","
	}
	if len(result) != 1 {
		result = result[:len(result)-1]
	}
	result += "}"
	return result, nil
}

func ToRecords(body string) ([]records.Record, error) {
	records := []records.Record{}
	err := json.Unmarshal([]byte(body), &records)
	return records, err
}

func ToId(body string) string {
	return body[1 : len(body)-1]
}
