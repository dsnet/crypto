// Copyright 2014, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// A high performance pseudo-random number generator.
package main

import "os"
import "io"
import "syscall"
import "runtime"
import "bitbucket.org/rawr/gorandom/rand"

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// Since rand.Reader never returns os.EOF, the only way io.Copy stops normally
// is if the os.Stdout pipe gets closed. Since this is expected behaviour,
// that error is ignored.
func errIgnorePipe(err error) error {
	if perr, ok := err.(*os.PathError); ok && perr.Err == syscall.EPIPE {
		return nil
	}
	return err
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.SetNumRoutines(runtime.NumCPU())
	_, err := io.Copy(os.Stdout, rand.Reader)
	errPanic(errIgnorePipe(err))
}
