// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

/*
This program must be on the most current zos release.

This program generates with data from

	//\'CEE.SCEELIB\(CELQS003\)\'

to output files:

	libvec_offsets.go (offset from libvec)

synopsis:

	  go run ./mklibvec.go

	or (with default flags)
	  go run mksyscall_zos_s390x.go -o libvec_offsets.go

	or if processed on a different platform
	  go run ./mklibvec.go -i CELQS003.txt
		in which CELQS003.txt is a text file copy of //\'CEE.SCEELIB\(CELQS003\)\'
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

var (
	sysnumfile = flag.String("o", "libvec_zos.go", "zos LE offsets output file in Go")
	testfile   = flag.String("i", "", "file for local validation")
)

// cmdLine returns this programs's commandline arguments
func cmdLine() string {
	_, fileName, _, _ := runtime.Caller(1)
	return "go run " + path.Base(fileName) + " -o " + *sysnumfile
}

func out(ch chan string, file io.ReadCloser) {
	defer file.Close()
	defer close(ch)
	rd := bufio.NewReader(file)
loop:
	for {
		str, err := rd.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Fatal("Read Error:", err)
			}
			break loop
		} else {
			ch <- str
		}
	}
}

type SO struct {
	Symbol string
	Offset int64
}

type SOList []SO

func (s SOList) Len() int      { return len(s) }
func (s SOList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SOList) Less(i, j int) bool {
	if s[i].Offset == s[j].Offset {
		return s[i].Symbol < s[j].Symbol
	}
	return s[i].Offset < s[j].Offset
}

// Param is function parameter
type Param struct {
	Name string
	Type string
}

// parseParamList parses parameter list and returns a slice of parameters
func parseParamList(list string) []string {
	list = strings.TrimSpace(list)
	if list == "" {
		return []string{}
	}
	return regexp.MustCompile(`\s*,\s*`).Split(list, -1)
}

// parseParam splits a parameter into name and type
func parseParam(p string) Param {
	ps := regexp.MustCompile(`^(\S*) (\S*)$`).FindStringSubmatch(p)
	if ps == nil {
		fmt.Fprintf(os.Stderr, "malformed parameter: %s\n", p)
		os.Exit(1)
	}
	return Param{ps[1], ps[2]}
}

func main() {
	flag.Parse()
	sidedeck := "//'CEE.SCEELIB(CELQS003)'"
	if *testfile != "" {
		sidedeck = *testfile
	}
	args := []string{"-u", sidedeck}
	cmd := exec.Command("/bin/cat", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		println("err stdout ")
		log.Fatal(err)
	}
	c1 := make(chan string)
	go out(c1, stdout)
	err2 := cmd.Start()
	if err2 != nil {
		log.Fatal(err2)
	}
	longest := 0
	outstanding := 1
	// IMPORT DATA64,CELQV003,'environ',001
	r1 := regexp.MustCompile("^ +IMPORT +CODE64,CELQV003,'([A-Za-z_][A-Za-z0-9_]*)',([0-9A-F][0-9A-F][0-9A-F]) *\n$")
	m := make(map[string]int64)
	for outstanding > 0 {
		select {
		case msg1, ok := <-c1:
			if ok {
				result := r1.FindStringSubmatch(msg1)
				if result != nil {
					if len(result) > 2 {
						symbol := "SYS_" + strings.ToUpper(result[1])
						offset, e1 := strconv.ParseInt(result[2], 16, 64)
						if e1 == nil {
							if len(symbol) > longest {
								longest = len(symbol)
							}
							m[symbol] = offset
						} else {
							fmt.Printf("ERROR %s\n", msg1)
						}
					}
				}
			} else {
				c1 = nil
				outstanding--
			}

		}
	}

	list := make(SOList, len(m))

	i := 0
	for k, v := range m {
		list[i] = SO{k, v}
		i++
	}
	sort.Sort(list)
	fmt.Printf("Writing %s\n", *sysnumfile)
	err = writesysnum(*sysnumfile, &list)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writesysnum %s %v\n", *sysnumfile, err)
		os.Exit(1)
	}
	err = gofmt(*sysnumfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gofmt %s %v\n", *sysnumfile, err)
		os.Exit(1)
	}

}

func writesysnum(file string, l *SOList) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	defer f.Close()
	defer w.Flush()
	fmt.Fprintf(w, `
//go:build zos
// +build zos

package utils

const (
`)
	for _, item := range *l {
		fmt.Fprintf(w, "    %-40s = 0x%X   // %d\n", item.Symbol, item.Offset, item.Offset)
	}
	fmt.Fprintf(w, `
)`)
	return nil
}
func gofmt(file string) error {
	cmd := exec.Command("gofmt", "-w", file)
	_, err := cmd.Output()

	if err != nil {
		return err
	}

	return nil
}
