package application

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalcHandler(t *testing.T) {
	testCasesBadRequest := []struct {
		name       string
		expression *strings.Reader
	}{
		{
			name:       "2+",
			expression: strings.NewReader(`{"expression": "2+"}`),
		},
		{
			name:       "2+3)",
			expression: strings.NewReader(`{"expression": "2+3)"}`),
		},
		{
			name:       "2/0",
			expression: strings.NewReader(`{"expression": "2/0"}`),
		},
		{
			name:       "2",
			expression: strings.NewReader(`{"expression": "2"}`),
		},
		{
			name:       "",
			expression: strings.NewReader(`{"expression": ""}`),
		},
	}

	for _, testCase := range testCasesBadRequest {
		req := httptest.NewRequest(http.MethodPost, "/", testCase.expression)
		w := httptest.NewRecorder()
		CalcHandler(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("wrong status code")
		}
	}

	testCasesOkResult := []struct {
		name           string
		expression     *strings.Reader
		expectedResult float64
	}{
		{
			name:       "2+2",
			expression: strings.NewReader(`{"expression": "2+2"}`),
		},
		{
			name:       "0/2",
			expression: strings.NewReader(`{"expression": "0/2"}`),
		},
	}

	for _, testCase := range testCasesOkResult {
		req := httptest.NewRequest(http.MethodPost, "/", testCase.expression)
		w := httptest.NewRecorder()
		CalcHandler(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("wrong status code")
		}
	}
}
