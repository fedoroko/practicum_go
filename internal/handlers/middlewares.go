package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// gzipWriter често пытался реализовать самописный middleware,
// текущая реализация отрабатывает локально, но не проходит тесты,
// видимо gzip.NewWriterLevel() жрет много ресурсов. Нужно реализовывать пулл,
// что я рад бы сделать, но боюсь не успею, посему отложил до следующего спринта,
// в текущей версии использую стороннее решение
type gzipWriter struct {
	http.ResponseWriter
	GZWriter    io.Writer
	writeHeader bool
	contentL    int
}

func (w *gzipWriter) WriteHeader(code int) {
	defer w.ResponseWriter.WriteHeader(code)
	if w.writeHeader {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
	} else {
		w.Header().Set("Content-length", fmt.Sprintf("%d", w.contentL))
	}
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	encodeIf := map[string]struct{}{
		"application/json": {},
		"text/plain":       {},
	}
	ct := w.Header().Get("Content-Type")
	_, ok := encodeIf[ct]
	l := len(b)
	if !ok || l < 50 {
		w.contentL = l
		w.WriteHeader(http.StatusOK)
		return w.ResponseWriter.Write(b)
	}

	w.writeHeader = true
	w.WriteHeader(http.StatusOK)
	return w.GZWriter.Write(b)

}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		// Видел, что тут стоит использовать пулл для врайтеров,
		// но вроде блокером не является, вынес пока в TO DO, реализаую позднее.
		// Кстати, заметил, что во всех либах пулл реализуют через глобальную переменную,
		// на первом код ревью были замечания касательно глобальных переменных,
		// прошу еще раз разъяснить, приемлимо ли их использование или нет.
		gz, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
		defer gz.Close()
		next.ServeHTTP(&gzipWriter{
			ResponseWriter: w,
			GZWriter:       gz,
		}, r)
	})
}
