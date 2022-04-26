package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fedoroko/practicum_go/internal/errrs"
	"github.com/fedoroko/practicum_go/internal/metrics"
	"github.com/fedoroko/practicum_go/internal/mocks"
)

type input struct {
	name  string
	mtype string
	value *float64
}

func Test_repoHandler_GetFunc(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}
	tests := []struct {
		name   string
		input  input
		output metrics.Metric
		err    error
		want   want
	}{
		{
			name: "positive test #1",
			input: input{
				name:  "Alloc",
				mtype: "gauge",
			},
			output: metrics.New(
				"Alloc",
				"gauge",
				1,
				0,
			),
			err: nil,
			want: want{
				code:        200,
				body:        "1",
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			input: input{
				name:  "Alloc",
				mtype: "int",
			},
			output: metrics.New(
				"Alloc",
				"int",
				0,
				0,
			),
			err: errrs.ThrowInvalidTypeError("int"),
			want: want{
				code:        501,
				body:        "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "wrong key",
			input: input{
				name:  "zAlloc",
				mtype: "gauge",
			},
			output: metrics.New(
				"zAlloc",
				"gauge",
				0,
				0,
			),
			err: errors.New("not found"),
			want: want{
				code:        404,
				body:        "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := mocks.NewMockRepository(ctrl)

	h := NewRepoHandler(db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := metrics.Raw(tt.input.mtype, tt.input.name)
			db.EXPECT().Get(m).Return(tt.output, tt.err)

			request := httptest.NewRequest(http.MethodGet, "/value/{type}/{name}", nil)
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.GetFunc)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.input.mtype)
			rctx.URLParams.Add("name", tt.input.name)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-type"))

			if tt.want.code == 200 {
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.want.body, string(resBody))
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
	var dummyOne float64 = 1
	tests := []struct {
		name   string
		input  input
		output input
		err    error
		body   string
		want   want
	}{
		{
			name: "positive test #1",
			output: input{
				name:  "Alloc",
				mtype: "gauge",
				value: &dummyOne,
			},
			err:  nil,
			body: "{\"id\":\"Alloc\",\"type\":\"counter\"}",
			want: want{
				code:        200,
				body:        "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}",
				contentType: "application/json",
			},
		},
		{
			name: "wrong type",
			output: input{
				name:  "Alloc",
				mtype: "int",
			},
			err:  errrs.ThrowInvalidTypeError("int"),
			body: "{\"id\":\"Alloc\",\"type\":\"int\"}",
			want: want{
				code:        501,
				body:        "Invalid type: int",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "wrong key",
			output: input{
				name:  "zlloc",
				mtype: "gauge",
			},
			err:  errors.New("not found"),
			body: "{\"id\":\"zlloc\",\"type\":\"gauge\"}",
			want: want{
				code:        404,
				body:        "not found",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := mocks.NewMockRepository(ctrl)

	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := metrics.FromJSON(bytes.NewBuffer([]byte(tt.body)))
			ret, _ := metrics.Raw(tt.output.mtype, tt.output.name)
			if tt.output.value != nil {
				ret.SetFloat64(*tt.output.value)
			}
			db.EXPECT().Get(m).Return(ret, tt.err)

			request := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.GetJSONFunc)

			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			if tt.want.code == 200 {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, tt.want.body, string(resBody))
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
	type txtInput struct {
		name  string
		mtype string
		value string
	}
	tests := []struct {
		name  string
		input txtInput
		err   error
		want  want
	}{
		{
			name: "positive test #1",
			input: txtInput{
				name:  "Alloc",
				mtype: "gauge",
				value: "1",
			},
			err: nil,
			want: want{
				code:        200,
				emptyBody:   true,
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			input: txtInput{
				name:  "Alloc",
				mtype: "int",
				value: "1",
			},
			err: errrs.ThrowInvalidTypeError("int"),
			want: want{
				code:        501,
				emptyBody:   false,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := mocks.NewMockRepository(ctrl)

	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := metrics.RawWithValue(tt.input.mtype, tt.input.name, tt.input.value)
			db.EXPECT().Set(m).Return(tt.err)

			request := httptest.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.UpdateFunc)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.input.mtype)
			rctx.URLParams.Add("name", tt.input.name)
			rctx.URLParams.Add("value", tt.input.value)
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

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

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
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
		err  error
		want want
	}{
		{
			name: "positive test #1",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":1}",
			err:  nil,
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "wrong type",
			body: "{\"id\":\"Alloc\",\"type\":\"int\",\"value\":1}",
			err:  errrs.ThrowInvalidTypeError("int"),
			want: want{
				code:        501,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "empty value",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\"}",
			err:  errors.New(""),
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "non numeric value",
			body: "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":\"int\"}",
			err:  errors.New(""),
			want: want{
				code:        400,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	db := mocks.NewMockRepository(ctrl)

	h := NewRepoHandler(db)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := metrics.FromJSON(bytes.NewBuffer([]byte(tt.body)))
			db.EXPECT().Set(m).Return(tt.err)
			request := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			hl := http.HandlerFunc(h.UpdateJSONFunc)

			rctx := chi.NewRouteContext()
			request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

			hl.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
