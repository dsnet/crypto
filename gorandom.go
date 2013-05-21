package main

import "os"
import "io"
import "bufio"
import "bytes"
import "runtime"
import "crypto/aes"
import "crypto/rand"
import "crypto/cipher"

// Number of chained blocks to generate with AES before re-seeding the key
// and CBC initialization vectors.
const NUM_CHAINED_BLOCKS = 1 << 20

func RandGen(comm chan io.Reader, quit chan bool) {
	data := make([]byte, aes.BlockSize*NUM_CHAINED_BLOCKS)
	for {
		// Generate a random new encryption key
		key := make([]byte, aes.BlockSize)
		if num, _ := rand.Read(key); num != aes.BlockSize {
			panic("Could not seed a random cryptographic key.")
		}

		// Create the encryption cipher
		aesCipher, err := aes.NewCipher(key)
		if err != nil {
			panic("Could not create a new AES cipher.")
		}

		// Generate a random initialization vector for cipher block chaining
		vector := make([]byte, aes.BlockSize)
		if num, _ := rand.Read(vector); num != aes.BlockSize {
			panic("Could not seed a random initialization vector.")
		}

		// Use CBC to generate pseudo-random data
		cbcCipher := cipher.NewCBCEncrypter(aesCipher, vector)
		cbcCipher.CryptBlocks(data, data)
		buf := bytes.NewBuffer(data)

		// Send data to main routine
		select {
		case comm <- buf:
			continue
		case <-quit:
			break
		}
	}
}

func main() {
	comm := make(chan io.Reader)
	quit := make(chan bool)

	// Spin off a great number of workers
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
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
