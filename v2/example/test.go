// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ibmruntimes/go-recordio/v2"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

const FIXED_PRODID_SIZE = 4
const FIXED_KEY_SIZE = 16
const FIXED_VAL_SIZE = 16

var stream = zosrecordio.RecordStream{}

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
	var key0[FIXED_KEY_SIZE]byte
	log.SetFlags(log.Lshortfile)
	buffbytes := utils.ConvertTypeToSlice(&buff)
	if len(os.Args) < 2 {
		fmt.Println("Provide an argument of the form \"//'HLQ.DBNAME.KEY.PATH'\"")
		return
	}
	Dname := os.Args[1]
	stream = zosrecordio.Fopen(Dname, "rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	stream = zosrecordio.Freopen(Dname,"rb+,type=record",stream)
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	copy(buff.key[:],key0[:])
	copy(buff.key[:], "KEY")
	copy(buff.val[:], "qrt")
	stream.Fwrite(buffbytes)
	stream.Fclose()
	stream = zosrecordio.Fopen(Dname,"rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	num = stream.Fread(buffbytes)
	eofp := stream.Feof()
	err := stream.Ferror()
	fmt.Println("--", num, string(buff.key[:]), string(buff.val[:]), eofp, err)
	myBigSlice = make([]byte, 100, 100)
	buffp, buffSize := utils.ConvertSliceToType[FixedHeader_T](myBigSlice)
	fmt.Println("The size of the struct is", buffSize)
	copy(buff.key[:],key0[:])
	copy(buffp.key[:], "KEY_AWAY")
	copy(buffp.val[:], "qrs")
	stream.Fwrite(myBigSlice[:buffSize])
	stream.Fclose()
	stream = zosrecordio.Fopen(Dname,"rb+,type=record")
	if stream.Nil() {
		log.Fatal("zero stream")
	}
	for {
		num = stream.Fread(myBigSlice)
		if stream.Feof() {
			break
		}
		fmt.Println("--", num, string(buffp.key[:]), string(buffp.val[:]))
	}
	fl := stream.Flocate([]byte("KEY_AWAY"), zosrecordio.Loc_KEY_EQ)
	fmt.Println(fl)
	if fl == 0 {
		fmt.Println("good flocate")
	}
	fmt.Println("zosrecordio.Flocate:", fl)
	num = stream.Fread(buffbytes)
	buff.val[0] = 'Y'
	buff.val[1] = 'X'
	numu = stream.Fupdate(buffbytes)
	fmt.Println(numu)
	numu = stream.Fclose()
	fmt.Println("close: ", numu)
	stream = zosrecordio.Fopen(Dname,"rb+,type=record")
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
