// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build zos

package recordio

import (
	"reflect"
	"runtime"
	"syscall"
	"unsafe"
)

const EOF = -1

// RecordStream holds a stream identifier
type RecordStream struct {
	s uintptr
}

// Used to determine if the stream has a valid idetifier
func (rs RecordStream) Nil() bool {
	return rs.s == 0
}

// Fopen takes a name and mode as byte slices and returns a RecordStream.
// Note that the strings encoded in thee slices must be null terminated.
// The utility function ConvertStringToSlice automatically null-terminates.
func Fopen(fname []byte, mode []byte) (rs RecordStream) {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x753<<4, //fopen
		[]uintptr{uintptr(unsafe.Pointer(&fname[0])),
			uintptr(unsafe.Pointer(&mode[0]))})
	rs.s = ret
	return rs
}

// Freopen behaves the same as Fopen, but takes a previously used RecordStream
func Freopen(fname []byte, mode []byte, rs RecordStream) (rso RecordStream) {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x754<<4, //freopen
		[]uintptr{uintptr(unsafe.Pointer(&fname[0])),
			uintptr(unsafe.Pointer(&mode[0])),
			rs.s})
	rso.s = ret
	return rso
}

// Flocate locates a record by Key, received as a byte slice
// Returns 0 if successful, otherwise EOF
func (rs RecordStream) Flocate(key []byte, options int) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x246<<4, //flocate
		[]uintptr{rs.s,
			uintptr(unsafe.Pointer(&key[0])),
			uintptr(len(key)), uintptr(options)})
	return int(ret)
}

// Fread reads a record.
// If the buffer is not big enough, the record will be truncated.
// The actual number of bytes read is returned
func  (rs RecordStream) Fread(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x78a<<4, //fread
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(1),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

// Fdelrec deletes the last read record
// Returns 0 if successful, otherwise non-zero
func  (rs RecordStream) Fdelrec() int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x247<<4, //fdelrec
		[]uintptr{rs.s})
	return int(ret)
}

// Feof returns the last set EOF flag value
func  (rs RecordStream) Feof() bool {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x04D<<4, //feof
		[]uintptr{rs.s})
	return int(ret) != 0
}

// Ferror returns the last set error value
func  (rs RecordStream) Ferror() error {
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
func  (rs RecordStream) Fupdate(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x0B5<<4, //fupdate
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

//Fwrite writes one record contained in buffer to the rs stream.
// It returns the number of bytes written.
// Note, teh size of the record is the size of the slice.
func  (rs RecordStream) Fwrite(buffer []byte) int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x78b<<4, //fwrite
		[]uintptr{uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(1),
			uintptr(len(buffer)),
			rs.s})
	return int(ret)
}

// Fclose closes the stream. Returns 0 if successful, otherwise EOF.
func  (rs RecordStream) Fclose() int {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x067<<4, //fclose
		[]uintptr{rs.s})
	return int(ret)
}

// ConvertStringToSlice copies the string into the given slice.
// Always includeds a null terminator in the copy.
// Returns a new empty slice if the string doesn't fit.
func ConvertStringToSlice(s string, bi []byte) (bo []byte) {
	size := len(s)
	if size < cap(bi) {
		copy(bi[:size], s)
		bi[size] = 0
		return bi[:size+1]
	} else {
		return []byte{}
	}
}

// ConvertStructToSlice returns a byte slice that shares storage
// with the incoming struct. The length of the slice will be
// exactly the size of the struct.
func ConvertStructToSlice(i interface{}) (slice []byte) {
	var size int
	var ptr unsafe.Pointer
	ptr = unsafe.Pointer(reflect.ValueOf(i).Elem().UnsafeAddr())
	size = int(reflect.ValueOf(i).Elem().Type().Size())
	data := (*(*[1<<31 - 1]byte)(ptr))[:size]
	return data
}

// ConvertSliceToStruc returns an interface that can be type asserted to be the
// same type as the incoming pointer to struct "i".
// The pointer that results from such a type assertion will share storage
// with the incoming byte slice "bi". The struct size is the second returned value.
// Thus the following sequence leaves buffp as a pointer to a FixedHeader struct
// that shares storage with myBigSlice[:buffSize]:
// var buffp *FixedHeader_T
// buffp_, buffSize := recordio.ConvertSliceToStruct(buffp, myBigSlice)
// buffp = buffp_.(*FixedHeader_T)
// Note: if the  incoming slice isn't big enough, it returns <nil, 0>.
func ConvertSliceToStruct(i interface{}, bi []byte) (interface{}, int) {
	bip := (&bi[0])
	bipp := &bip
	result := reflect.NewAt(reflect.ValueOf(i).Type(), unsafe.Pointer(bipp)).Elem().Interface()
	size := int(reflect.ValueOf(result).Elem().Type().Size())
	if size > len(bi) {
		return nil, 0
	} else {
		return result, size
	}
}
