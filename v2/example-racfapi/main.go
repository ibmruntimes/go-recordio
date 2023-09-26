package main

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

type PLIST struct {
	list          [15]uint32
	Workarea      [128]uint64
	SafRcAlet     int32
	Rc            int32
	RacfRcAlet    int32
	RacfRc        int32
	RacfRsnAlet   int32
	RacfRsnRc     int32
	FuncAlet      int32
	Func          int16
	OptionWord    int32
	RacfUseridLen byte
	RacfUserid    [8]byte
	CertLen       byte
	Cert          [4096]byte
	AppIdLen      int16
	AppId         [246]byte
	DistNameLen   int16
	DistName      [246]byte
	RegNameLen    int16
	RegName       [255]byte
}

func main() {
	// Set up PLIST
	siz := (int((reflect.TypeOf((*PLIST)(nil)).Elem()).Size()))
	plist31 := (*PLIST)(unsafe.Pointer(utils.Malloc31(siz)))
	plist31.list[0] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.Workarea)))
	plist31.list[1] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.SafRcAlet)))
	plist31.list[2] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.Rc)))
	plist31.list[3] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RacfRcAlet)))
	plist31.list[4] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RacfRc)))
	plist31.list[5] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RacfRsnAlet)))
	plist31.list[6] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RacfRsnRc)))
	plist31.list[7] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.FuncAlet)))
	plist31.list[8] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.Func)))
	plist31.list[9] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.OptionWord)))
	plist31.list[10] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RacfUseridLen)))
	plist31.list[11] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.CertLen)))
	plist31.list[12] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.AppIdLen)))
	plist31.list[13] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.DistNameLen)))
	plist31.list[14] = uint32(0x7fffffff & uintptr(unsafe.Pointer(&plist31.RegNameLen)))
	plist31.list[14] |= uint32(0x80000000)

	// LOAD IRRSIM64
	mod := utils.LoadMod("IRRSIM00")
	if uintptr(unsafe.Pointer(mod)) != 0 {
		plist31.Func = 0x3
		copy(plist31.RacfUserid[:], "IBMUSER ")
		utils.AtoE(plist31.RacfUserid[:])
		plist31.RacfUseridLen = 7 // length of userid
		RC := mod.Call(uintptr(unsafe.Pointer(plist31)))
		if RC == 0 {
			fmt.Printf("SafRC %d RacfRc %d Reason %d\n", plist31.Rc, plist31.RacfRc, plist31.RacfRsnRc)
		} else {
			fmt.Printf("Call rc=0x%x\n", RC)
		}
		// FREE MODULE
		mod.Free()
	} else {
		fmt.Printf("Failed to load IRRSIM00\n")
	}
	// FREE PARM STORAGE
	utils.Free(unsafe.Pointer(plist31))
}
