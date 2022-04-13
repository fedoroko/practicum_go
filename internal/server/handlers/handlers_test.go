package handlers

import (
	"bytes"
	"context"
	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_repoHandler_GetFunc(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}
	type params struct {
		Type  string
		Name  string
		Value string
	}
	tests := []struct {
		name      string
		urlParams params
		want      want
	}{
		{
			name: "positive test #1",
			urlParams: params{
				Type: "gauge",
				Name: "Alloc",
			},
			want: want{
				code:        200,
				body:        "1",
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			urlParams: params{
				Type: "int",
				Name: "Alloc",
			},
			want: want{
				code:        501,
				body:        "Invalid type: int",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "wrong key",
			urlParams: params{
				Type: "gauge",
				Name: "int",
			},
			want: want{
				code:        404,
				body:        "not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	db := storage.New(config.NewServerConfig().Default())
	_ = db.Set(
		storage.RawWithValue("gauge", "Alloc", "1"),
	)
	defer db.Close()
	h := NewRepoHandler(db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/value/{type}/{name}", nil)
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.GetFunc)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.urlParams.Type)
			rctx.URLParams.Add("name", tt.urlParams.Name)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if strings.TrimSpace(string(resBody)) != tt.want.body {
				t.Errorf("Expected body \"%s\", got \"%s\"", tt.want.body, w.Body.String())
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func Test_repoHandler_GetJSONFunc(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\"}",
			want: want{
				code:        200,
				body:        "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}",
				contentType: "application/json",
			},
		},
		{
			name: "wrong type",
			body: "{\"id\":\"Alloc\",\"type\":\"int\"}",
			want: want{
				code:        501,
				body:        "Invalid type: int",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "wrong key",
			body: "{\"id\":\"zlloc\",\"type\":\"gauge\"}",
			want: want{
				code:        404,
				body:        "not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	db := storage.New(config.NewServerConfig().Default())
	_ = db.Set(
		storage.RawWithValue("gauge", "Alloc", "1"),
	)
	defer db.Close()
	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.GetJSONFunc)

			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if strings.TrimSpace(string(resBody)) != tt.want.body {
				t.Errorf("Expected body \"%s\", got \"%s\"", tt.want.body, w.Body.String())
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func Test_repoHandler_UpdateFunc(t *testing.T) {
	type want struct {
		code        int
		emptyBody   bool
		contentType string
	}
	type params struct {
		Type  string
		Name  string
		Value string
	}
	tests := []struct {
		name      string
		urlParams params
		want      want
	}{
		{
			name: "positive test #1",
			urlParams: params{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "1",
			},
			want: want{
				code:        200,
				emptyBody:   true,
				contentType: "text/plain",
			},
		},
		{
			name: "positive test #2",
			urlParams: params{
				Type:  "counter",
				Name:  "alloc",
				Value: "1",
			},
			want: want{
				code:        200,
				emptyBody:   true,
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			urlParams: params{
				Type:  "int",
				Name:  "Alloc",
				Value: "1",
			},
			want: want{
				code:        501,
				emptyBody:   false,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "empty value",
			urlParams: params{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "",
			},
			want: want{
				code:        400,
				emptyBody:   false,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "non numeric value",
			urlParams: params{
				Type:  "gauge",
				Name:  "Alloc",
				Value: "none",
			},
			want: want{
				code:        400,
				emptyBody:   false,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	db := storage.New(config.NewServerConfig().Default())
	defer db.Close()
	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.UpdateFunc)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.urlParams.Type)
			rctx.URLParams.Add("name", tt.urlParams.Name)
			rctx.URLParams.Add("value", tt.urlParams.Value)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if tt.want.emptyBody {
				if string(resBody) != "" {
					t.Errorf("Expected empty body, got %s", w.Body.String())
				}
			} else {
				if string(resBody) == "" {
					t.Errorf("Expected non empty body, got empty")
				}
			}

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func Test_repoHandler_UpdateJSONFunc(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "positive test #1",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			body: "{\"id\":\"Alloc\",\"type\":\"int\",\"value\":1}",
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "empty value",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\"}",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "non numeric value",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":\"int\"}",
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	db := storage.New(config.NewServerConfig().Default())
	defer db.Close()
	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.UpdateJSONFunc)

			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()

			if res.Header.Get("Content-Type") != tt.want.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
