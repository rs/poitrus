package overlay

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func Handler(h http.Handler, origin string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Host = origin
		},
		ModifyResponse: func(r *http.Response) error {
			if r.StatusCode == 404 {
				r.Header = http.Header{}
				r.Body.Close()
				r.Body = ioutil.NopCloser(&bytes.Buffer{})
			}
			return nil
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			h.ServeHTTP(w, r)
			return
		}
		wr := httptest.NewRecorder()
		wr.Body = &bytes.Buffer{}
		h.ServeHTTP(wr, r)
		if wr.Code == 404 {
			proxy.ServeHTTP(w, r)
			return
		}
		hdr := w.Header()
		for k, v := range wr.Header() {
			hdr[k] = v
		}
		w.WriteHeader(wr.Code)
		w.Write(wr.Body.Bytes())
		return
	})
}
