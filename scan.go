// Copyright 2015 Giulio Iotti. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"io"
)

type scanner struct {
	*bufio.Scanner
}

func newScanner(r io.Reader) *scanner {
	sc := &scanner{Scanner: bufio.NewScanner(r)}
	sc.Scanner.Split(bufio.ScanWords)
	return sc
}

type consumer struct {
	graph         *graph
	firsts        []string
	reader        io.Reader
	buf           bytes.Buffer
	npars, nwords int64
}

func newConsumer(r io.Reader) *consumer {
	return &consumer{
		graph:  newGraph(),
		firsts: make([]string, 0),
		reader: r,
	}
}

func (c *consumer) ingest() error {
	var lastVal string
	s := newScanner(&c.buf)
	for s.Scan() {
		val := s.Text()
		c.graph.add(val)
		if lastVal != "" {
			c.graph.link(lastVal, val)
		} else {
			c.firsts = append(c.firsts, val)
		}
		lastVal = val
		c.nwords++
	}
	c.npars++
	return s.Err()
}

func (c *consumer) avgLen() int64 {
	return c.nwords / c.npars
}

func (c *consumer) readAll() error {
	scanner := bufio.NewScanner(c.reader)
	c.graph.Lock()
	defer c.graph.Unlock()
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		c.buf.Write(scanner.Bytes())
		if err := c.ingest(); err != nil {
			return err
		}
	}
	return scanner.Err()
}
