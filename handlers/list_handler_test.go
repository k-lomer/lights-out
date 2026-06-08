package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the list handler with default params returns some outages.
func Test_ListHandler_Basic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	res := httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.Greater(t, len(outages), 0)
}

// Test the list handler with a small page size returns that number of outages.
func Test_ListHandler_PageSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "2")
	res := httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.Equal(t, len(outages), 2)
}

// Test the list handler with a different page index returns different outages.
func Test_ListHandler_PageIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageIndex", "0")
	res := httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages1 := decodeOutages(t, res.Body)

	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageIndex", "1")
	res = httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages2 := decodeOutages(t, res.Body)

	assert.Equal(t, len(outages1), len(outages2))
	for i := range 2 {
		assert.NotEqual(t, outages1[i], outages2[i])
	}
}

// Test the list handler with page size 0 returns outages from all DNOs.
func Test_ListHandler_AllOutages(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	res := httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	checkDnoOutages(t, outages)
}
