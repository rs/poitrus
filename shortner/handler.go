package shortner

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	DB *DB
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
	target, found, err := h.DB.Get(path)
	if err != nil {
		log.Printf("GET %s: db error: %v", path, err)
		sendError(w, http.StatusInternalServerError)
		return
	}
	if !found {
		log.Printf("GET %s: not found", path)
		sendError(w, http.StatusNotFound)
		return
	}
	log.Printf("GET %s => %s", path, target)
	w.Header().Set("Location", target)
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.WriteHeader(http.StatusMovedPermanently)
}

func parseURL(target string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(target))
	if err != nil {
		return "", err
	}
	if !parsed.IsAbs() {
		return "", errors.New("not absolute")
	}
	if parsed.Host == "" {
		return "", errors.New("empty hostname")
	}
	switch parsed.Scheme {
	case "http", "https":
	default:
		return "", errors.New("unsupported scheme")
	}
	return parsed.String(), nil
}

func (h Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("PUT %s: cannot read body: %v", path, err)
		sendError(w, http.StatusInternalServerError)
		return
	}
	target, err := parseURL(string(body))
	if err != nil {
		sendError(w, http.StatusBadRequest)
		log.Printf("PUT %s: invalid target: %v", path, err)
		return
	}
	if exists, err := h.DB.Set(r.URL.Path, target); err != nil {
		sendError(w, http.StatusInternalServerError)
		log.Printf("PUT %s: db error: %v", path, err)
		return
	} else if exists {
		sendError(w, http.StatusConflict)
		log.Printf("PUT %s: conflict: %v", path, err)
		return
	}
	log.Printf("PUT %s => %s", path, target)
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	found, err := h.DB.Delete(path)
	if err != nil {
		log.Printf("DELETE %s: db error: %v", path, err)
		sendError(w, http.StatusInternalServerError)
		return
	}
	if !found {
		log.Printf("DELETE %s: not found", path)
		sendError(w, http.StatusNotFound)
		return
	}
	log.Printf("DELETE %s", path)
	w.WriteHeader(http.StatusNoContent)
}

func sendError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}
