// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/ibmruntimes/go-recordio/v2"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func main() {
	fh := zosrecordio.Fopen("//'SYS1.MACLIB(EXCP)'", "rb, lrecl=80, blksize=80, recfm=fb, type=record")
	if fh.Nil() {
		utils.Perror()
		os.Exit(1)
	}
	defer fh.Fclose()
	var line [80]byte
	bytes := fh.Fread(line[:])
	cnt := 0
	for bytes > 0 {
		utils.EtoA(line[:bytes])
		fmt.Printf("%s\n", string(line[:bytes]))
		cnt++
		bytes = fh.Fread(line[:])
	}
	if fh.Ferror() != nil {
		utils.Perror()
	}
}
