// Copyright 2015 Giulio Iotti. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"log"
	"mime/multipart"
	"net/http"
)

var errInvalidMethod = errors.New("only POST or GET handled")

type server struct {
	host   string
	maxLen int64
	cons   *consumer
}

func newServer(host string, c *consumer) *server {
	return &server{
		host:   host,
		cons:   c,
		maxLen: 2 * 1024 * 1024,
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		log.Print("invalid request to ", r.URL.Path)
		return
	}
	var err error
	switch r.Method {
	case "POST":
		err = s.post(w, r)
	case "GET":
		err = s.get(w, r)
	default:
		err = errInvalidMethod
	}
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) postFile(f *multipart.FileHeader) error {
	fh, err := f.Open()
	if err != nil {
		return err
	}
	defer fh.Close()
	return s.cons.readAll(fh)
}

func (s *server) post(w http.ResponseWriter, r *http.Request) error {
	s.cons.graph.Lock()
	defer s.cons.graph.Unlock()
	if err := r.ParseMultipartForm(s.maxLen); err != nil {
		return err
	}
	defer r.MultipartForm.RemoveAll()
	for _, fg := range r.MultipartForm.File {
		for _, file := range fg {
			s.postFile(file)
		}
	}
	// TODO: Redirect to GET?
	return nil
}

func (s *server) get(w http.ResponseWriter, r *http.Request) error {
	s.cons.graph.RLock()
	defer s.cons.graph.RUnlock()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	walker := newWalker(w, s.cons.graph)
	walker.seed(walker.any(s.cons.firsts))
	return walker.walk()
}

func (s *server) serve() error {
	return http.ListenAndServe(s.host, s)
}
