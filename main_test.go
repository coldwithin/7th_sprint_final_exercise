package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		count int
		want  int
	}{
		{count: 0, want: 1}, // при count = 0 в ответе пустая строка, которую Split разбивает в слайс с длиной 1
		{count: 1, want: 1},
		{count: 2, want: 2},
		{count: 100, want: min(len(cafeList["moscow"]), 100)},
	}

	for _, test := range tests {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=moscow&count=%d", test.count), nil)

		handler.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)

		result := strings.Split(strings.TrimSpace(resp.Body.String()), ",")

		if test.count == 0 { // при count = 0 можно проверить на пустой ответ
			assert.Equal(t, "", result[0])
		}

		assert.Len(t, result, test.want)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	tests := []struct {
		search    string
		wantCount int
	}{
		{search: "фасоль", wantCount: 1}, // при отсутствии значения в ответе пустая строка, которую Split разбивает в слайс с длиной 1
		{search: "кофе", wantCount: 2},
		{search: "вилка", wantCount: 1},
	}

	for _, test := range tests {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/cafe?city=moscow&search=%s", test.search), nil)

		handler.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)

		resultStr := strings.TrimSpace(resp.Body.String())

		assert.Len(t, strings.Split(resultStr, ","), test.wantCount)
		assert.True(
			t,
			resultStr == "" && test.wantCount == 1 || strings.Contains(strings.ToLower(resultStr), strings.ToLower(test.search)),
		)
	}
}
