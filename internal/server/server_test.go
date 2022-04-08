package server

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	r := router()
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, body := testRequest(t, ts, "POST", "/update/gauge/alloc/1", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "", body)

	resp, _ = testRequest(t, ts, "POST", "/update/gauge/alloc/none", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, body = testRequest(t, ts, "GET", "/value/gauge/alloc", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "1", body)

	resp, _ = testRequest(t, ts, "GET", "/value/int/alloc", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)

	resp, _ = testRequest(t, ts, "GET", "/", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	inp := bytes.NewBuffer([]byte("{\"id\":\"alloc\",\"type\":\"gauge\",\"value\":393728}"))
	resp, _ = testRequest(t, ts, "POST", "/update", inp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	inp = bytes.NewBuffer([]byte("{\"id\":\"alloc\",\"type\":\"gauge\"}"))
	resp, body = testRequest(t, ts, "POST", "/value", inp)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"id\":\"alloc\",\"type\":\"gauge\",\"value\":393728}", body)
}
