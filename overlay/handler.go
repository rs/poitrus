package overlay

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
)

func Handler(h http.Handler, origin string) http.Handler {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL.Scheme = "https"
			r.Out.URL.Host = origin
			log.Printf("Proxying to %s", r.Out.URL.String())
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		ModifyResponse: func(r *http.Response) error {
			if r.StatusCode == 404 {
				r.Header = http.Header{}
				r.Body.Close()
				r.Body = io.NopCloser(&bytes.Buffer{})
			}
			return nil
		},
		ErrorLog: log.Default(),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			h.ServeHTTP(w, r)
			return
		}
		wr := httptest.NewRecorder()
		wr.Body = &bytes.Buffer{}
		h.ServeHTTP(wr, r)
		log.Printf("GET %s: status %d", r.URL.Path, wr.Code)
		if wr.Code == 404 {
			proxy.ServeHTTP(w, r)
			return
		}
		hdr := w.Header()
		for k, v := range wr.Header() {
			hdr[k] = v
		}
		w.WriteHeader(wr.Code)
		_, _ = w.Write(wr.Body.Bytes())
	})
}
