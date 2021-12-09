// +build zos
package utils

import (
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
