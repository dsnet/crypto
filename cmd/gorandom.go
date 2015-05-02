// Copyright 2014, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// A high performance pseudo-random number generator.
package main

import "os"
import "io"
import "fmt"
import "syscall"
import "runtime"
import "unsafe"
import "github.com/ogier/pflag"
import "bitbucket.org/rawr/gorandom/rand"

// Check if the given file descriptor writes to a terminal.
func isatty(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TCGETS, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func main() {
	// Basic user configuration variables
	count := pflag.Int64P("count", "n", -1, "Number of random bytes to generate")
	force := pflag.BoolP("force", "f", false, "Force output to terminal")
	procs := pflag.IntP("procs", "p", runtime.NumCPU(), "Maximum number of concurrent workers")
	pflag.Parse()

	if !(*force) && isatty(syscall.Stdout) {
		fmt.Fprintf(os.Stderr, "Random data not written to terminal.\n\n")
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
	var err error
	if (*count) < 0 {
		_, err = io.Copy(os.Stdout, rand.Reader)
	} else {
		_, err = io.CopyN(os.Stdout, rand.Reader, *count)
	}
	if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.EPIPE {
		err = nil // Expected error is for the sink to close the pipe
	} else if err != nil {
		panic(err)
	}
}
