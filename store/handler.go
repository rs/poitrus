package store

import (
	"io"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	Store Store
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.handleGet(w, r)
	case "PUT":
		h.handlePut(w, r)
	case "DELETE":
		h.handleDelete(w, r)
	default:
		sendError(w, http.StatusMethodNotAllowed)
	}
}

func (h Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	e, err := h.Store.Get(path)
	if err != nil {
		sc := http.StatusInternalServerError
		if err == ErrNotFound {
			sc = http.StatusNotFound
		}
		log.Printf("GET %s: store error: %v (status %d)", path, err, sc)
		sendError(w, sc)
		return
	}
	defer e.Body.Close()
	log.Printf("GET %s", path)
	sc := http.StatusOK
	for k, v := range e.Header {
		switch k {
		case "Status":
			if i, err := strconv.Atoi(v[0]); err == nil && sc >= 100 && sc < 600 {
				sc = i
			}
			continue
		case "Location":
			if sc == http.StatusOK {
				sc = http.StatusPermanentRedirect
			}
		}
		w.Header()[k] = v
	}
	w.WriteHeader(sc)
	io.Copy(w, e.Body)
}

func (h Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	hdr := r.Header
	hdr.Del("User-Agent")
	hdr.Del("Accept")
	hdr.Del("Accept-Encoding")
	hdr.Del("Date")
	hdr.Del("Transfer-Encoding")
	if err := h.Store.Set(r.URL.Path, &Entry{hdr, r.Body}); err != nil {
		sc := http.StatusInternalServerError
		if err == ErrExists {
			sc = http.StatusConflict
		}
		sendError(w, sc)
		log.Printf("PUT %s: store error: %v (status %d)", path, err, sc)
		return
	}
	log.Printf("PUT %s", path)
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	err := h.Store.Delete(path)
	if err != nil {
		sc := http.StatusInternalServerError
		if err == ErrNotFound {
			sc = http.StatusNotFound
		}
		log.Printf("DELETE %s: db error: %v (status %d)", path, err, sc)
		sendError(w, sc)
		return
	}
	log.Printf("DELETE %s", path)
	w.WriteHeader(http.StatusNoContent)
}

func sendError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
