// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var line [132]byte
	bytes := zosrecordio.Stdin().Fread(line[:])
	cnt := 0
	for bytes > 0 {
		zosrecordio.Stdout().Fwrite(line[:bytes])
		zosrecordio.Stdout().Fwrite([]byte("\x15"))
		cnt++
		bytes = zosrecordio.Stdin().Fread(line[:])
	}
}
