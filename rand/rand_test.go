// Copyright 2015, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package rand

import "io"
import "bytes"
import "compress/flate"
import "testing"
import "github.com/stretchr/testify/assert"

func TestRead(t *testing.T) {
	var n int = 1 << 24 // 16 MiB
	if testing.Short() {
		n = 1 << 16 // 64 KiB
	}

	d := make([]byte, n)
	n, err := io.ReadFull(Reader, d)
	assert.Equal(t, len(d), n)
	assert.Nil(t, err)

	var b bytes.Buffer
	z, _ := flate.NewWriter(&b, 5)
	z.Write(d)
	z.Close()
	assert.True(t, b.Len() >= len(d)*99/100)
}

func TestReadEmpty(t *testing.T) {
	n, err := Reader.Read(make([]byte, 0))
	assert.Equal(t, 0, n)
	assert.Nil(t, err)

	n, err = Reader.Read(nil)
	assert.Equal(t, 0, n)
	assert.Nil(t, err)
}
