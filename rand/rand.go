// Copyright 2014, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package rand

import "io"
import "sync"
import "crypto/aes"
import "crypto/rand"
import "crypto/cipher"

var reader = newCrypter()
var Reader io.Reader = reader

type crypter struct {
	dataChan chan []byte
	monChan  chan mondata
	buffer   []byte
	lock     sync.Mutex
}

type mondata struct {
	num int
	ret chan int
}

func newCrypter() *crypter {
	crypt := new(crypter)
	crypt.dataChan = make(chan []byte, 1)
	crypt.monChan = make(chan mondata, 1)
	crypt.monChan <- mondata{1, nil}
	go crypt.monitor()
	return crypt
}

func (c *crypter) Read(buf []byte) (n int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.buffer) == 0 {
		c.buffer = <-c.dataChan
	}
	n = copy(buf, c.buffer)
	c.buffer = c.buffer[n:]

	return n, nil
}

func (c *crypter) monitor() {
	var genQuits []chan bool
	var md mondata
	for ok := true; ok; {
		select {
		case md, ok = <-c.monChan:
			numPre := len(genQuits)

			// Start routines
			for len(genQuits) < md.num {
				quit := make(chan bool)
				genQuits = append(genQuits, quit)
				go c.generator(quit)
			}

			// End routines
			for _, quit := range genQuits[md.num:] {
				close(quit)
			}
			genQuits = genQuits[:md.num]

			if md.ret != nil {
				md.ret <- numPre
			}
		}
	}
}

func (c *crypter) generator(quit chan bool) {
	const minBlocks = 256
	const maxBlocks = 65536

	// Create a new encryption cipher with random key
	key := make([]byte, aes.BlockSize)
	if _, err := rand.Read(key); err != nil {
		panic("Could not seed a random cryptographic key.")
	}
	aesCipher, _ := aes.NewCipher(key)

	// Create a new CBC generator with a random initialization vector
	vector := make([]byte, aes.BlockSize)
	if _, err := rand.Read(vector); err != nil {
		panic("Could not seed a random initialization vector.")
	}
	cbcCipher := cipher.NewCBCEncrypter(aesCipher, vector)

	numBlocks := minBlocks
	data := make([]byte, aes.BlockSize*numBlocks)
	for live := true; live; {
		// Generate pseudo-random data by encrypting in CBC mode
		cbcCipher.CryptBlocks(data, data)

		// Grow the blocksize to be more efficient
		dataCopy := data
		if numBlocks < maxBlocks {
			data = append(data, make([]byte, aes.BlockSize*numBlocks)...)
			numBlocks *= 2
		} else {
			dataCopy = append([]byte(nil), data...)
		}

		select {
		case c.dataChan <- dataCopy:
			continue
		case live = <-quit:
			break
		}
	}
}

// Read is a helper function that calls Reader.Read using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(buf []byte) (n int, err error) {
	return io.ReadFull(Reader, buf)
}

// Sets the number of routines that will generate pseudo-random data and returns
// the previous setting. If n < 1, it does not change the current setting.
// By default, the number of routines starts off at 1.
func SetNumRoutines(num int) int {
	ret := make(chan int)
	reader.monChan <- mondata{num, ret}
	return <-ret
}
