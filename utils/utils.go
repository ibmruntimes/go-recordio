// +build zos
package utils

import (
	"reflect"
	"runtime"
	"unsafe"
)

func EtoA(record []byte) {
	sz := len(record)
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x6e3<<4, // __e2a_l
		[]uintptr{uintptr(unsafe.Pointer(&record[0])), uintptr(sz)})
}

func AtoE(record []byte) {
	sz := len(record)
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x741<<4, // __a2e_l
		[]uintptr{uintptr(unsafe.Pointer(&record[0])), uintptr(sz)})
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
// buffp_, buffSize := zosrecordio.ConvertSliceToStruct(buffp, myBigSlice)
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

// The equivalent of the perror() function which prints an error string on the "current" errno.
//
func Perror() {
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x712<<4, //perror()
		[]uintptr{})
}

// The equivalent of the dup2() function which duplicates a file descriptor
//
func Dup2(oldfd uintptr, newfd uintptr) uintptr {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+0x183<<4,
		[]uintptr{oldfd, newfd})
	return ret
}
