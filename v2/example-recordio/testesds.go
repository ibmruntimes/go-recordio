// Copyright 2026 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// testesds demonstrates using this package on a VSAM ESDS
// (entry-sequenced data set), which can be treated like a
// sequential data set with record I/O:
//
//   - Fwrite appends records at the end of the ESDS
//   - Fread reads records sequentially
//   - Flocate(key, Loc_RBA_EQ) does random access by RBA
//     (pass the RBA as a right-justified binary value, 4 or 8 bytes)
//   - Fupdate updates the last read record in place
//   - Fdelrec is NOT supported on ESDS and returns an error
//     (record deletion is only valid for KSDS/RRDS)
//
// Note: when defining the cluster with IDCAMS, code NONINDEXED
// explicitly, otherwise IDCAMS defaults to a KSDS with KEYS(64,0):
//
//   DEFINE CLUSTER (NAME(HLQ.ESDS) NONINDEXED LOG(NONE) UNIQUE -
//      TRACKS(1,1) RECSZ(80,80))

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ibmruntimes/go-recordio/v2"
)

const RECSIZE = 80

// dump reads and prints all records sequentially
func dump(stream zosrecordio.RecordStream, label string) {
	buf := make([]byte, RECSIZE)
	num := 0
	for {
		n := stream.Fread(buf)
		if n <= 0 {
			break
		}
		num++
		fmt.Printf("%s record %d (%d bytes): %q...\n", label, num, n, string(buf[:12]))
	}
	fmt.Printf("%s: %d records, feof=%v err=%v\n", label, num, stream.Feof(), stream.Ferror())
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) < 2 {
		fmt.Println("Provide an argument of the form \"//'HLQ.ESDS'\"")
		return
	}
	Dname := os.Args[1]

	// Append three fixed-length records
	stream := zosrecordio.Fopen(Dname, "rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	for _, val := range []string{"AAAA", "BBBB", "CCCC"} {
		buf := make([]byte, RECSIZE)
		copy(buf, val)
		for i := len(val); i < RECSIZE; i++ {
			buf[i] = '.'
		}
		n := stream.Fwrite(buf)
		fmt.Println("Fwrite:", n, stream.Ferror())
	}
	stream.Fclose()

	// Read them back sequentially
	stream = zosrecordio.Fopen(Dname, "rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	dump(stream, "sequential")

	// Locate the second record by RBA (records are RECSIZE bytes,
	// so the second record starts at RBA 80) and update it in place
	rba := make([]byte, 8)
	rba[7] = RECSIZE
	fl := stream.Flocate(rba, zosrecordio.Loc_RBA_EQ)
	fmt.Println("Flocate RBA=80:", fl, stream.Ferror())
	if fl == 0 {
		buf := make([]byte, RECSIZE)
		n := stream.Fread(buf)
		fmt.Printf("read after Flocate: %d %q...\n", n, string(buf[:12]))
		copy(buf[4:8], "ZZZZ")
		nu := stream.Fupdate(buf)
		fmt.Println("Fupdate:", nu, stream.Ferror())
	}
	stream.Fclose()

	// Verify the update, then demonstrate that Fdelrec is not
	// supported for ESDS (it returns an error and the record stays)
	stream = zosrecordio.Fopen(Dname, "rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	dump(stream, "after update")
	rba[7] = 2 * RECSIZE
	fl = stream.Flocate(rba, zosrecordio.Loc_RBA_EQ)
	fmt.Println("Flocate RBA=160:", fl, stream.Ferror())
	if fl == 0 {
		fd := stream.Fdelrec()
		fmt.Println("Fdelrec (expected to fail on ESDS):", fd, stream.Ferror())
	}
	stream.Fclose()

	// Final dump: all three records still present
	stream = zosrecordio.Fopen(Dname, "rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	dump(stream, "final")
	stream.Fclose()
}
