// Copyright 2015 Giulio Iotti. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"log"
	"os"
)

func readFile(c *consumer, name string) error {
	var r io.Reader
	if name == "-" {
		r = os.Stdin
	} else {
		fh, err := os.Open(name)
		if err != nil {
			return err
		}
		defer fh.Close()
		r = fh
	}
	return c.readAll(r)
}

func main() {
	var (
		lines = flag.Int64("lines", 0, "Number `N` of lines to output. Default is the same as lines in input.")
		http  = flag.String("http", "", "Listen to H (format: '[HOST]:PORT') for HTTP GET and POST requests.")
	)
	flag.Parse()
	args := flag.Args()
	// By default, read from standard input.
	if len(args) == 0 && *http == "" {
		args = append(args, "-")
	}
	var readSuccess bool
	c := newConsumer()
	for i := range args {
		if err := readFile(c, args[i]); err != nil {
			log.Print(err)
		}
		readSuccess = true
	}
	// If there is no data, we could still get it via HTTP.
	if *http != "" {
		server := newServer(*http, c)
		log.Print("server listening on ", *http)
		log.Fatal(server.serve())
		// HTTP server won't proceed from here.
	}
	// If we read no data, just quit.
	if !readSuccess {
		return
	}
	if *lines == 0 {
		*lines = c.npars
	}
	walker := newWalker(os.Stdout, c.graph)
	for i := int64(0); i < *lines; i++ {
		walker.seed(walker.any(c.firsts))
		walker.walk()
	}
}
