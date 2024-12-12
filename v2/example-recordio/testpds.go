// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	zosrecordio "github.com/ibmruntimes/go-recordio/v2"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func main() {
	signal.Ignore(syscall.SIGIOERR)
	path := flag.String("i", "//'SYS1.MACLIB(EXCP)'", "Data set member e.g. \"//'SYS1.MACLIB(EXCP)'\"")
	options := flag.String("t", "rb, lrecl=80, blksize=80, recfm=fb, type=record", "fopen open options e.g. \"rb, lrecl=80, blksize=80, recfm=fb, type=record\"")
	flag.Parse()
	fh := zosrecordio.Fopen(*path, *options)
	if fh.Nil() {
		utils.Perror()
		os.Exit(1)
	}
	defer fh.Fclose()
	var line [32768]byte
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
