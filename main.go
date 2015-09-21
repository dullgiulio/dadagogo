// Copyright 2015 Giulio Iotti. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		lines = flag.Int64("lines", 0, "Number `N` of lines to output. Default is the same as lines in input.")
	)
	flag.Parse()
	c := newConsumer(os.Stdin)
	if err := c.readAll(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
		return
	}
	if *lines == 0 {
		*lines = c.npars
	}
	c.graph.RLock()
	defer c.graph.RUnlock()
	walker := newWalker(os.Stdout, c.graph)
	for i := int64(0); i < *lines; i++ {
		walker.seed(walker.any(c.firsts))
		walker.random()
	}
}
