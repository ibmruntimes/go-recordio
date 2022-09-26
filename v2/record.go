// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build zos

package zosrecordio

import (
	"runtime"
	"syscall"
	"unsafe"
)

const EOF = -1

type LocOptions int // Flocate Options

const (
	Loc_RBA_EQ LocOptions = iota
	Loc_KEY_FIRST
	Loc_KEY_LAST
	Loc_KEY_EQ
	Loc_KEY_EQ_BWD
	Loc_KEY_GE
	Loc_RBA_EQ_BWD
)

// RecordStream holds a stream identifier
type RecordStream struct {
	s uintptr
}

// Used to determine if the stream has a valid idetifier
func (rs RecordStream) Nil() bool {
	return rs.s == 0
}

// Fopen takes a name and mode as strings  and returns a RecordStream.
func Fopen(fname string, mode string) (rs RecordStream) {
	fnameBytes := []byte(fname + "\x00")
	modeBytes := []byte(mode + "\x00")
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x753<<4, //fopen
		[]uintptr{uintptr(unsafe.Pointer(&fnameBytes[0])),
			uintptr(unsafe.Pointer(&modeBytes[0]))})
	rs.s = ret
	return rs
}

// Freopen behaves the same as Fopen, but takes a previously used RecordStream
func Freopen(fname string, mode string, rs RecordStream) (rso RecordStream) {
	fnameBytes := []byte(fname + "\x00")
	modeBytes := []byte(mode + "\x00")
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x754<<4, //freopen
		[]uintptr{uintptr(unsafe.Pointer(&fnameBytes[0])),
			uintptr(unsafe.Pointer(&modeBytes[0])),
			rs.s})
	rso.s = ret
	return rso
}

// Flocate locates a record by Key, received as a byte slice
// Returns 0 if successful, otherwise EOF
func (rs RecordStream) Flocate(key []byte, options LocOptions) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x246<<4, //flocate
		[]uintptr{rs.s,
			uintptr(unsafe.Pointer(&key[0])),
			uintptr(len(key)), uintptr(options)})
	return int(ret)
}

// Fread reads a record.
// If the buffer is not big enough, the record will be truncated.
// The actual number of bytes read is returned
func (rs RecordStream) Fread(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x78a<<4, //fread
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(1),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

// Fdelrec deletes the last read record
// Returns 0 if successful, otherwise non-zero
func (rs RecordStream) Fdelrec() int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x247<<4, //fdelrec
		[]uintptr{rs.s})
	return int(ret)
}

// Feof returns the last set EOF flag value
func (rs RecordStream) Feof() bool {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x04D<<4, //feof
		[]uintptr{rs.s})
	return int(ret) != 0
}

// Ferror returns the last set error value
func (rs RecordStream) Ferror() error {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x04A<<4, //feof
		[]uintptr{rs.s})
	if int(ret) == 0 {
		return nil
	} else {
		return syscall.Errno(ret)
	}
}

// Fupdate updates the last read record to be the new record value in buffer
// It returns the size of the updated record
func (rs RecordStream) Fupdate(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x0B5<<4, //fupdate
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

//Fwrite writes one record contained in buffer to the rs stream.
// It returns the number of bytes written.
// Note, teh size of the record is the size of the slice.
func (rs RecordStream) Fwrite(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x78b<<4, //fwrite
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(1),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

// Fclose closes the stream. Returns 0 if successful, otherwise EOF.
func (rs RecordStream) Fclose() int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x067<<4, //fclose
		[]uintptr{rs.s})
	return int(ret)
}

// The equivalent of STDIN
func Stdin() (rs RecordStream) {
	rs.s = stdio_filep(0)
	return rs
}

// The equivalent of STDOUT
func Stdout() (rs RecordStream) {
	rs.s = stdio_filep(1)
	return rs
}

// The equivalent of STDERR
func Stderr() (rs RecordStream) {
	rs.s = stdio_filep(2)
	return rs
}

func stdio_filep(fd int32) uintptr {
	return uintptr(*(*uint64)(unsafe.Pointer(uintptr(*(*uint64)(
		unsafe.Pointer(uintptr(*(*uint64)(unsafe.Pointer(uintptr(
			uint64(*(*uint32)(unsafe.Pointer(uintptr(1208)))) + 80))) +
			uint64((fd+2)<<3))))))))
}
