package main

import "os"
import "io"
import "fmt"
import "path"
import "flag"
import "bufio"
import "bytes"
import "runtime"
import "crypto/aes"
import "crypto/rand"
import "crypto/cipher"

const DESCRIPTION = `
Another pseudo-random number generator (PRNG). This was designed to be a faster
alternative to the /dev/urandom device built into most Linux kernels. This PRNG
is based on the Advanced Encryption Standard (AES) cipher operating in
cipher-block chaining (CBC) mode. Both the AES key and CBC initialization vector
will be seeded from the operating system's cryptographically secure PRNG.

This implementation is capable of spawning multiple routines that generate
pseudo-random data in parallel. By default, this generator will spawn off a
number of routines equal to the number of logical cores in an attempt to
maximize the output. If the output is not being consumed as fast as data is
being generated, routines will block. As a result, CPU utilization may go down,
allowing other processes to run.
`

var NumJobs, NumBlocks int

func RandGen(comm chan io.Reader, quit chan bool) {
	data := make([]byte, aes.BlockSize*NumBlocks)
	for {
		// Create a new encryption cipher with random key
		key := make([]byte, aes.BlockSize)
		if num, _ := io.ReadFull(rand.Reader, key); num != aes.BlockSize {
			panic("Could not seed a random cryptographic key.")
		}
		aesCipher, _ := aes.NewCipher(key)

		// Create a new CBC generator with a random initialization vector
		vector := make([]byte, aes.BlockSize)
		if num, _ := io.ReadFull(rand.Reader, vector); num != aes.BlockSize {
			panic("Could not seed a random initialization vector.")
		}
		cbcCipher := cipher.NewCBCEncrypter(aesCipher, vector)

		// Generate pseudo-random data by encrypting in CBC mode
		cbcCipher.CryptBlocks(data, data)

		// Send data to main routine
		select {
		case comm <- bytes.NewBuffer(data):
			continue
		case <-quit:
			break
		}
	}
}

func main() {
	comm := make(chan io.Reader)
	quit := make(chan bool)

	// Command-line parser
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fs.Usage = func() {
		cmd := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", cmd)
		fmt.Fprintln(os.Stderr, DESCRIPTION)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Report '%s' bugs to joetsai@digital-static.net\n", cmd)
	}
	fs.IntVar(&NumJobs, "jobs", runtime.NumCPU(), "Number of Go-routines jobs to spin off.")
	fs.IntVar(&NumBlocks, "blocks", (1 << 16), "Each routine will generate this many pseudo-random blocks before re-seeding.")
	fs.Parse(os.Args[1:])
	if NumJobs <= 0 {
		fmt.Fprintln(os.Stderr, "Number of jobs must be positive.")
		os.Exit(1)
	}
	if NumBlocks <= 0 {
		fmt.Fprintln(os.Stderr, "Number of blocks to generate must be positive.")
		os.Exit(1)
	}

	// Spin off a great number of workers
	runtime.GOMAXPROCS(NumJobs)
	for i := 0; i < NumJobs; i++ {
		go RandGen(comm, quit)
	}

	// Continually read data from workers
	dstBuf := bufio.NewWriter(os.Stdout)
	for {
		srcBuf := <-comm
		if _, err := io.Copy(dstBuf, srcBuf); err != nil {
			close(quit)
			break
		}
	}
}
