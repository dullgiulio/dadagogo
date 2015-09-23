// Copyright 2015 Giulio Iotti. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"io"
	"math/rand"
	"sync"
	"time"
)

type weight int64

type node map[string]weight

func makeNode(s string) node {
	return make(map[string]weight)
}

type graph struct {
	sync.RWMutex
	dict map[string]node
}

func newGraph() *graph {
	return &graph{
		dict: make(map[string]node),
	}
}

func (g *graph) add(s string) {
	if _, ok := g.dict[s]; !ok {
		g.dict[s] = makeNode(s)
	}
}

func (g *graph) link(from, to string) bool {
	nfrom, ok := g.dict[from]
	if !ok {
		return false
	}
	nfrom[to]++
	return true
}

type walker struct {
	graph  *graph
	rnd    *rand.Rand
	writer io.Writer
	val    string
}

func newWalker(w io.Writer, g *graph) *walker {
	return &walker{
		graph:  g,
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
		writer: w,
	}
}

func (w *walker) any(s []string) string {
	if len(s) == 0 {
		return ""
	}
	n := w.rnd.Int63n(int64(len(s)))
	return s[n]
}

func (w *walker) hasNext(nd node) bool {
	return len(nd) != 0
}

func (w *walker) weightedRand(nd node) string {
	// Sum all the counts.
	var total weight
	for _, n := range nd {
		total += n
	}
	// TODO: check for overflows?
	// Get a number n, 0 <= n < total
	n := weight(w.rnd.Int63n(int64(total)))
	// Find which node i this number corresponds to in the sum.
	total = weight(0)
	for k, v := range nd {
		ntotal := weight(total + v + 1)
		if total <= n && n < ntotal {
			return k
		}
		total = ntotal
	}
	panic("not reachable")
}

func (w *walker) seed(s string) {
	w.val = s
}

func (w *walker) walk() error {
	if w.val == "" {
		return errors.New("walker needs to be seeded")
	}
	if _, err := io.WriteString(w.writer, w.val); err != nil {
		return err
	}
	if _, err := io.WriteString(w.writer, " "); err != nil {
		return err
	}
	for {
		if !w.hasNext(w.graph.dict[w.val]) {
			break
		}
		w.val = w.weightedRand(w.graph.dict[w.val])
		if _, err := io.WriteString(w.writer, w.val); err != nil {
			return err
		}
		if _, err := io.WriteString(w.writer, " "); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w.writer, "\n"); err != nil {
		return err
	}
	w.val = ""
	return nil
}
