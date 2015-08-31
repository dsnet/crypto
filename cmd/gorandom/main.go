// Copyright 2014, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// A high performance pseudo-random number generator.
package main

import "os"
import "io"
import "fmt"
import "math"
import "syscall"
import "runtime"
import "github.com/ogier/pflag"
import "golang.org/x/crypto/ssh/terminal"
import "github.com/dsnet/gorandom/rand"
import "github.com/dsnet/golib/strconv"

func main() {
	// Basic user configuration variables
	force := pflag.BoolP("force", "f", false, "Force output to terminal.")
	count := pflag.StringP("count", "n", "+Inf", "Number of random bytes to generate.")
	procs := pflag.IntP("procs", "p", runtime.NumCPU(), "Maximum number of concurrent workers.")
	pflag.Parse()

	if !(*force) && terminal.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintf(os.Stderr, "Random data not written to terminal.\n\n")
		pflag.Usage()
		os.Exit(1)
	}
	cnt, err := strconv.ParsePrefix(*count, strconv.AutoParse)
	if err != nil || math.IsNaN(cnt) {
		fmt.Fprintf(os.Stderr, "Number of bytes to generate is invalid.\n\n")
		pflag.Usage()
		os.Exit(1)
	}
	if (*procs) < 1 {
		fmt.Fprintf(os.Stderr, "Number of workers must be positive.\n\n")
		pflag.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(*procs)
	rand.SetNumRoutines(*procs)

	// Copy random data to stdout
	if int64(cnt) < 0 || math.IsInf(cnt, 0) {
		_, err = io.Copy(os.Stdout, rand.Reader)
	} else {
		_, err = io.CopyN(os.Stdout, rand.Reader, int64(cnt))
	}
	if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.EPIPE {
		err = nil // Expected error is for the sink to close the pipe
	} else if err != nil {
		panic(err)
	}
}
