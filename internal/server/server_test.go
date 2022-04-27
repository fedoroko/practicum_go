package server

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fedoroko/practicum_go/internal/errrs"
	"github.com/fedoroko/practicum_go/internal/metrics"
	"github.com/fedoroko/practicum_go/internal/mocks"
	"github.com/fedoroko/practicum_go/internal/storage"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var db storage.Repository
	tdb := mocks.NewMockRepository(ctrl)
	m, _ := metrics.RawWithValue("gauge", "Alloc", "1")
	tdb.EXPECT().Set(m).Return(nil)

	m, _ = metrics.RawWithValue("gauge", "Alloc", "none")
	tdb.EXPECT().Set(m).Return(errors.New(""))

	m, _ = metrics.Raw("gauge", "Alloc")
	ret, _ := metrics.RawWithValue("gauge", "Alloc", "1")
	tdb.EXPECT().Get(m).Return(ret, nil)

	m, _ = metrics.Raw("int", "Alloc")
	ret, _ = metrics.RawWithValue("int", "Alloc", "1")
	tdb.EXPECT().Get(m).Return(ret, errrs.ThrowInvalidTypeError("int"))

	tdb.EXPECT().List().Return([]metrics.Metric{}, nil)

	m, _ = metrics.FromJSON(bytes.NewBuffer([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":393728}")))
	tdb.EXPECT().Set(m).Return(nil)

	m, _ = metrics.FromJSON(bytes.NewBuffer([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}")))
	ret, _ = metrics.RawWithValue("gauge", "Alloc", "1")
	tdb.EXPECT().Get(m).Return(ret, nil)

	db = tdb
	r := router(&db)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, body := testRequest(t, ts, "POST", "/update/gauge/Alloc/1", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "", body)

	resp, _ = testRequest(t, ts, "POST", "/update/gauge/Alloc/none", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, body = testRequest(t, ts, "GET", "/value/gauge/Alloc", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "1", body)

	resp, _ = testRequest(t, ts, "GET", "/value/int/Alloc", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	resp, _ = testRequest(t, ts, "GET", "/", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	inp := bytes.NewBuffer([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":393728}"))
	resp, _ = testRequest(t, ts, "POST", "/update", inp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	inp = bytes.NewBuffer([]byte("{\"id\":\"Alloc\",\"type\":\"gauge\"}"))
	resp, body = testRequest(t, ts, "POST", "/value", inp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}", body)
}
