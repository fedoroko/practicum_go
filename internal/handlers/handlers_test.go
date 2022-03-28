package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateFunc(t *testing.T) {
	type want struct {
        code        int
        emptyBody 	bool    
        contentType string
    }
    tests := []struct {
        name string
        url string
        want want
    }{
        {
            name: "positive test #1",
            url: "/update/gauge/Alloc/1",
            want: want{
            	code: 200,
            	emptyBody: true,
            	contentType: "text/plain",
            },
        },
        {
            name: "wrong path #1",
            url: "/update/gauge/Alloc/1/",
            want: want{
            	code: 400,
            	emptyBody: false,
            	contentType: "text/plain; charset=utf-8",
            },
        },
        {
            name: "wrong path #2",
            url: "/update/gauge/Alloc",
            want: want{
            	code: 400,
            	emptyBody: false,
            	contentType: "text/plain; charset=utf-8",
            },
        },
        {
            name: "empty value",
            url: "/update/gauge/Alloc/",
            want: want{
            	code: 400,
            	emptyBody: false,
            	contentType: "text/plain; charset=utf-8",
            },
        },
        {
            name: "wrong type",
            url: "/update/error/Alloc/1",
            want: want{
            	code: 400,
            	emptyBody: false,
            	contentType: "text/plain; charset=utf-8",
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
        	request := httptest.NewRequest(http.MethodPost, tt.url, nil)
        	w := httptest.NewRecorder()
        	h := http.HandlerFunc(UpdateFunc)

        	h.ServeHTTP(w, request)
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

