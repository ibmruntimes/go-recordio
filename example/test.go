// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.ibm.com/open-z/go-recordio"
	"os"
)

const __KEY_EQ = 3
const FIXED_PRODID_SIZE = 4
const FIXED_KEY_SIZE = 16
const FIXED_VAL_SIZE = 16

var stream = recordio.RecordStream{}

type FixedHeader_T struct {
	inactive      byte
	reserved      byte
	prodID        [FIXED_PRODID_SIZE]byte
	key           [FIXED_KEY_SIZE]byte
	val           [FIXED_VAL_SIZE]byte
	prodIDXOffset uint16
	prodIDXLen    uint16
	keyXOffset    uint16
	keyXLen       uint16
	valXOffset    uint16
	valXLen       uint16
	filterXOffset uint16
	filterXLen    uint16
}

var buff = FixedHeader_T{}

func main() {
	var num int
	var numu int
	buffbytes := recordio.ConvertStructToSlice(&buff)
	myBigSlice := make([]byte, 100, 100)
	mySmallSlice := make([]byte, 20, 20)
	if len(os.Args) < 2 {
		fmt.Println("Provide an argument of the form \"//'HLQ.DBNAME.KEY.PATH'\"")
		return
	}
	Dname := os.Args[1]
	sliceBig := recordio.ConvertStringToSlice(Dname, myBigSlice)
	sliceSmall := recordio.ConvertStringToSlice("rb+,type=record", mySmallSlice)
	stream = recordio.Fopen(sliceBig, sliceSmall)
	if !stream.Nil() {
		fmt.Println("nonzero stream")
	}
	stream = recordio.Freopen(sliceBig, sliceSmall, stream)
	copy(buff.key[:], "KEY")
	copy(buff.val[:], "qrt")
	stream.Fwrite(buffbytes)
	stream.Fclose()
	stream = recordio.Fopen(sliceBig, sliceSmall)
	num = stream.Fread(buffbytes)
	eofp := stream.Feof()
	err := stream.Ferror()
	fmt.Println("--", num, string(buff.key[:]), string(buff.val[:]), eofp, err)
	myBigSlice = make([]byte, 100, 100)
	var buffp *FixedHeader_T
	buffp_, buffSize := recordio.ConvertSliceToStruct(buffp, myBigSlice)
	buffp = buffp_.(*FixedHeader_T)
	fmt.Println("The size of the struct is", buffSize)
	copy(buffp.key[:], "KEY_AWAY")
	copy(buffp.val[:], "qrs")
	stream.Fwrite(myBigSlice[:buffSize])
	stream.Fclose()
	stream = recordio.Fopen(sliceBig, sliceSmall)
	for {
		num = stream.Fread(myBigSlice)
		if stream.Feof() {
			break
		}
		fmt.Println("--", num, string(buffp.key[:]), string(buffp.val[:]))
	}
	fl := stream.Flocate([]byte("KEY_AWAY"), __KEY_EQ)
	fmt.Println(fl)
	if fl == 0 {
		fmt.Println("good flocate")
	}
	fmt.Println("recordio.Flocate:", fl)
	num = stream.Fread(buffbytes)
	buff.val[0] = 'Y'
	buff.val[1] = 'X'
	numu = stream.Fupdate(buffbytes)
	fmt.Println(numu)
	numu = stream.Fclose()
	fmt.Println("close: ", numu)
	stream = recordio.Fopen(sliceBig, sliceSmall)
	if !stream.Nil() {
		fmt.Println("nonzero stream")
	}
	for {
		num = stream.Fread(myBigSlice)
		if stream.Feof() {
			break
		}
		fmt.Println("--", num, string(buffp.key[:]), string(buffp.val[:]))
	}
}
