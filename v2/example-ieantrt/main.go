package main

import (
	"fmt"
	"reflect"
	"unsafe"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

type PLIST struct {
	list     [4]uint32
	toklevel uint32
	tokname  [16]byte
	tokvalue [16]byte
	rc       int32
}

func main() {
	// Set up PLIST
	siz := (int((reflect.TypeOf((*PLIST)(nil)).Elem()).Size()))
	plist31 := (*PLIST)(unsafe.Pointer(utils.Malloc31(siz)))
	plist31.toklevel = 0x04 // TOKLEVEL DC A(IEANT_SYSTEM_LEVEL)  = 0x04
	copy(plist31.tokname[:], "EZBSYSLOGD      ")
	utils.AtoE(plist31.tokname[:])
	copy(plist31.tokvalue[:], "                ")
	utils.AtoE(plist31.tokvalue[:])
	plist31.rc = 0
	plist31.list[0] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.toklevel)))
	plist31.list[1] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.tokname[0])))
	plist31.list[2] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.tokvalue[0])))
	plist31.list[3] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.rc)))
	plist31.list[3] |= uint32(0x80000000)
	syslogd := false

	// LOAD IEANTRT
	mod := utils.LoadMod("IEANTRT")
	if uintptr(unsafe.Pointer(mod)) != 0 {
		// CALL IEANTRT
		RC := mod.Call(uintptr(unsafe.Pointer(plist31)))
		if RC == 0 {
			if plist31.rc == 0 {
				fmt.Printf("TOKVALUE %v\n", plist31.tokvalue)
				syslogd = true
			} else {
				fmt.Printf("TOKRC 0x%x\n", plist31.rc)
			}
		} else {
			fmt.Printf("Call rc=0x%x\n", RC)
		}
		// FREE MODULE
		mod.Free()
	} else {
		fmt.Printf("Failed to load IEANTRT\n")
	}
	// FREE PARM STORAGE
	utils.Free(unsafe.Pointer(plist31))
	if syslogd {
		fmt.Printf("SYSLOGD is ACTIVE\n")
	} else {
		fmt.Printf("SYSLOGD is NOT ACTIVE\n")
	}
}
