package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/k-lomer/lights-out/cache"
	"github.com/k-lomer/lights-out/model"
	"github.com/stretchr/testify/assert"
)

// Test the list handler with default params returns some outages.
func Test_ListHandler_Basic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages)
}

// Test the list handler with a small page size returns that number of outages.
func Test_ListHandler_PageSize(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "2")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.Len(t, outages, 2)
}

// Test the list handler with a different page index returns different outages.
func Test_ListHandler_PageIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageIndex", "0")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages1 := decodeOutages(t, res.Body)

	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageIndex", "1")
	res = httptest.NewRecorder()

	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
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

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	checkDnoOutages(t, outages, model.AllDnoList[:])
}

// Test the postcode filter.
func Test_ListHandler_Postcodes(t *testing.T) {
	// Get all outages.
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	totalOutagesCount := len(outages)
	checkDnoOutages(t, outages, model.AllDnoList[:])

	// Get all outages for the first postcode.
	postcode := outages[0].Postcodes[0]
	req = httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, "postcodes", string(postcode))
	res = httptest.NewRecorder()

	lh.ServeHTTP(res, req)
	requireStatus(t, res.Code, http.StatusOK)
	outages = decodeOutages(t, res.Body)
	postcodeOutagesCount := len(outages)
	assert.Less(t, postcodeOutagesCount, totalOutagesCount)
	for _, o := range outages {
		assert.Equal(t, postcode, o.Postcodes[0])
	}
}

// Test the postcode filter when no postcodes match.
func Test_ListHandler_PostcodesNoMatches(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, "postcodes", "X00XX")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.Equal(t, 0, len(outages))
}

// Test the error case for invalid postcodes.
func Test_ListHandler_PostcodesInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, "postcodes", "XYZ")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusBadRequest)
}

// Test DNO selection.
func Test_ListHandler_DnoSelection(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, string(model.DnoEnergyNorthWest), "true")
	addQueryParams(req, string(model.DnoNationalGridDistribution), "true")
	addQueryParams(req, string(model.DnoNorthernPowergrid), "false")
	addQueryParams(req, string(model.DnoSPEnergy), "false")
	addQueryParams(req, string(model.DnoSse), "false")
	addQueryParams(req, string(model.DnoUKPowerNetwork), "false")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	checkDnoOutages(t, outages, []model.Dno{model.DnoEnergyNorthWest, model.DnoNationalGridDistribution})
}

// Test the status filter defaults to only active outages.
func Test_ListHandler_StatusDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages)
	for _, o := range outages {
		assert.Equal(t, model.StatusActive, o.Status)
	}
}

// Test selecting future outages returns only future outages.
func Test_ListHandler_StatusFuture(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, string(model.StatusActive), "false")
	addQueryParams(req, string(model.StatusFuture), "true")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages)
	for _, o := range outages {
		assert.Equal(t, model.StatusFuture, o.Status)
	}
}

// Test selecting resolved outages returns only resolved outages.
func Test_ListHandler_StatusResolved(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, string(model.StatusActive), "false")
	addQueryParams(req, string(model.StatusResolved), "true")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages)
	for _, o := range outages {
		assert.Equal(t, model.StatusResolved, o.Status)
	}
}

// Test selecting all statuses returns outages of every status.
func Test_ListHandler_StatusAll(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, string(model.StatusActive), "true")
	addQueryParams(req, string(model.StatusFuture), "true")
	addQueryParams(req, string(model.StatusResolved), "true")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)

	seen := map[model.Status]bool{}
	for _, o := range outages {
		seen[o.Status] = true
	}
	assert.True(t, seen[model.StatusActive])
	assert.True(t, seen[model.StatusFuture])
	assert.True(t, seen[model.StatusResolved])
}

// Test selecting future and resolved returns only those statuses, not active.
func Test_ListHandler_StatusFutureAndResolved(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, "pageSize", "0")
	addQueryParams(req, string(model.StatusActive), "false")
	addQueryParams(req, string(model.StatusFuture), "true")
	addQueryParams(req, string(model.StatusResolved), "true")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages := decodeOutages(t, res.Body)

	seen := map[model.Status]bool{}
	for _, o := range outages {
		assert.NotEqual(t, model.StatusActive, o.Status)
		seen[o.Status] = true
	}
	assert.True(t, seen[model.StatusFuture])
	assert.True(t, seen[model.StatusResolved])
}

// Test an invalid status value returns a bad request.
func Test_ListHandler_StatusInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, string(model.StatusActive), "maybe")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusBadRequest)
}

// Test targeting no statuses returns a bad request.
func Test_ListHandler_StatusNoneTargeted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	addQueryParams(req, string(model.StatusActive), "false")
	addQueryParams(req, string(model.StatusFuture), "false")
	addQueryParams(req, string(model.StatusResolved), "false")
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusBadRequest)
}

func Test_ListHandler_Caching(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), cache.MakeOutageCache(time.Minute))
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages1 := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages1)

	lh.ServeHTTP(res, req)
	requireStatus(t, res.Code, http.StatusOK)
	outages2 := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages2)

	assert.Equal(t, outages1, outages2)
}

func Test_ListHandler_NoCaching(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/list", nil)
	res := httptest.NewRecorder()

	lh := NewListHandler(NewTestDnoClients(), nil)
	lh.ServeHTTP(res, req)

	requireStatus(t, res.Code, http.StatusOK)
	outages1 := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages1)

	lh.ServeHTTP(res, req)
	requireStatus(t, res.Code, http.StatusOK)
	outages2 := decodeOutages(t, res.Body)
	assert.NotEmpty(t, outages2)

	assert.NotEqual(t, outages1, outages2)
}
