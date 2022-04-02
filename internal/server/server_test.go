package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	r := router()
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, body := testRequest(t, ts, "POST", "/update/gauge/alloc/1")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "", body)

	resp, _ = testRequest(t, ts, "POST", "/update/gauge/alloc/none")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, body = testRequest(t, ts, "GET", "/value/gauge/alloc")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "1", body)

	resp, _ = testRequest(t, ts, "GET", "/value/int/alloc")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	resp, _ = testRequest(t, ts, "GET", "/")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
