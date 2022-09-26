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

// ConvertTypeToSlice returns a byte slice that shares storage
// with the incoming object of type T. The length of the slice
// will be exactly the size of the struct.
// Note: it isn't necessary to explicitly pass in the type T,
// as it can be inferred via the argument i.
func ConvertTypeToSlice[T any](i T) (slice []byte) {
	var size int
	var ptr unsafe.Pointer
	ptr = unsafe.Pointer(reflect.ValueOf(i).Elem().UnsafeAddr())
	size = int(reflect.ValueOf(i).Elem().Type().Size())
	data := (*(*[1<<31 - 1]byte)(ptr))[:size]
	return data
}

// ConvertSliceToType returns a pointer to an object of type T.
// The pointer  will share storage with the incoming byte slice "bi". 
// The struct size is the second returned value.
// If the  incoming slice isn't big enough, it returns <nil, 0>.
func ConvertSliceToType[T1 any](bi []byte) (*T1, int) {
	bip := (&bi[0])
	bipp := &bip
	var t1p *T1
	t1pr := reflect.NewAt(reflect.ValueOf(t1p).Type(), unsafe.Pointer(bipp)).Elem().Interface()
	size := int(unsafe.Sizeof(*(t1pr.(*T1))))
	return t1pr.(*T1), size
}

// ConvertSliceToTypes is the same as / ConvertSliceToType, but with
// two objects of types T1 and T2, allocated contiguously in the slice
func ConvertSliceToTypes[T1,T2 any](bi []byte) (*T1, *T2) {
	bip := (&bi[0])
	bipp := &bip
	var t1p *T1
	t1pr := reflect.NewAt(reflect.ValueOf(t1p).Type(), unsafe.Pointer(bipp)).Elem().Interface()
	size1 := unsafe.Sizeof(*(t1pr.(*T1)))
	t2pr := unsafe.Pointer(uintptr(unsafe.Pointer(t1pr.(*T1))) + size1)
	return t1pr.(*T1), (*T2)(t2pr)
}

// ConvertSliceToTypes is the same as / ConvertSliceToType, but with
// three objects of types T1, T2, and T3, allocated contiguously in the slice
func ConvertSliceToTypes3[T1, T2, T3 any](bi []byte) (*T1, *T2, *T3) {
	bip := (&bi[0])
	bipp := &bip
	var t1p *T1
	t1pr := reflect.NewAt(reflect.ValueOf(t1p).Type(), unsafe.Pointer(bipp)).Elem().Interface()
	size1 := unsafe.Sizeof(*(t1pr.(*T1)))
	t2pr := unsafe.Pointer(uintptr(unsafe.Pointer(t1pr.(*T1))) + size1)
	t2prb := (*T2)(t2pr)
	size2 := unsafe.Sizeof(*(t2prb))
	t3pr := unsafe.Pointer(uintptr(unsafe.Pointer(t2prb)) + size2)
	return t1pr.(*T1), (*T2)(t2pr), (*T3)(t3pr)
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
