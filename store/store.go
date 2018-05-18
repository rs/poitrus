package store

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrExists        = errors.New("already exists")
	ErrInvalidFormat = errors.New("invalid format")
)

type Store struct {
	Root string
}

type Entry struct {
	Header http.Header
	Body   io.ReadCloser
}

type bufFile struct {
	*bufio.Reader
	f *os.File
}

func (bf bufFile) Close() error {
	return bf.f.Close()
}

func (s Store) path(path string) string {
	return filepath.Join(s.Root, fmt.Sprintf("%x", sha1.Sum([]byte(path))))
}

func (s Store) Get(path string) (*Entry, error) {
	p := s.path(path)
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	h := http.Header{}
	buf := bufio.NewReader(f)
	for {
		l, err := buf.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(l) == 1 {
			break
		}
		if idx := bytes.IndexByte(l, ':'); idx != -1 {
			h.Set(string(l[:idx]), string(bytes.TrimSpace(l[idx+1:len(l)-1])))
		} else {
			h.Set(string(l[:idx]), "")
		}
	}
	return &Entry{
		Header: h,
		Body:   bufFile{buf, f},
	}, nil
}

func (s Store) Set(path string, e *Entry) (err error) {
	p := s.path(path)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		if os.IsExist(err) {
			err = ErrExists
		}
		return err
	}
	for k, vs := range e.Header {
		for _, v := range vs {
			if strings.IndexByte(k, '\n') != -1 || strings.IndexByte(v, '\n') != -1 {
				f.Close()
				return ErrInvalidFormat
			}
			if _, err := f.WriteString(k + ": " + v + "\n"); err != nil {
				return err
			}
		}
	}
	if _, err := f.Write([]byte{'\n'}); err != nil {
		return err
	}
	if _, err := io.Copy(f, e.Body); err != nil {
		return err
	}
	return f.Close()
}

func (s Store) Delete(path string) error {
	p := s.path(path)
	err := os.Remove(p)
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	return err
}
