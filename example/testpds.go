// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.ibm.com/open-z/recordio"
	"github.ibm.com/open-z/recordio/utils"
)

func PrintError() {
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x712<<4, //perror()
		[]uintptr{})
}

func main() {
	fh := recordio.Fopen([]byte("//'SYS1.MACLIB(EXCP)'\x00"), []byte("rb, lrecl=80, blksize=80, recfm=fb, type=record\x00"))
	if fh.Nil() {
		PrintError()
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
		PrintError()
	}
}
