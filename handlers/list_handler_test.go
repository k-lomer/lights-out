package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_ListHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	res := httptest.NewRecorder()

	ListHandler(res, req)

	assertStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	checkDnoOutages(t, outages)
}
