package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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
	const city = "moscow"
	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, min(100, len(cafeList[city]))},
	}
	for _, tc := range requests {
		count := strconv.Itoa(tc.count)
		t.Run("count="+count, func(t *testing.T) {
			cafes := getCafes(t, url.Values{"city": {city}, "count": {count}})
			assert.Len(t, cafes, tc.want)
		})
	}
}

func TestCafeSearch(t *testing.T) {
	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
		{"КоФе", 2},
	}
	for _, tc := range requests {
		t.Run("search="+tc.search, func(t *testing.T) {
			cafes := getCafes(t, url.Values{"city": {"moscow"}, "search": {tc.search}})
			assert.Len(t, cafes, tc.wantCount)
			search := strings.ToLower(tc.search)
			for _, cafe := range cafes {
				assert.True(t, strings.Contains(strings.ToLower(cafe), search),
					"кафе %q должно содержать подстроку %q", cafe, tc.search)
			}
		})
	}
}

func getCafes(t *testing.T, query url.Values) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/cafe?"+query.Encode(), nil)
	response := httptest.NewRecorder()
	mainHandle(response, req)
	require.Equal(t, http.StatusOK, response.Code)

	body := response.Body.String()
	if body == "" {
		return nil
	}
	return strings.Split(body, ",")
}
