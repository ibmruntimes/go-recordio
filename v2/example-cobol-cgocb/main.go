package main

/*
#cgo LDFLAGS: XD2DLL.x
#include <stdlib.h>
#include <stdio.h>

typedef void (*callback_t)(char*, int);
extern int XDUMP2C(callback_t cb, char* str, int len);
extern void CblCallback(char * str, int len);

*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

//export CblCallback
func CblCallback(cstr *C.char, len C.int) {
	utils.ThreadAsciiMode() 
	str := C.GoString(cstr)
        fmt.Printf("\n\n\n\nCall from COBOL: %*s\n", len, str)
	runtime.HexDump(uintptr(unsafe.Pointer(cstr)), uintptr(len))
	utils.ThreadEbcdicMode() 
}

func main() {
	str := "hello world"
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
        cb := unsafe.Pointer(C.CblCallback)
	runtime.LockOSThread()   // lock thread
	utils.ThreadEbcdicMode() // COBOL must run in ebcdic
	ret := C.XDUMP2C((C.callback_t)(cb), cstr, C.int(len(str)))
	utils.ThreadAsciiMode()  // restore
	runtime.UnlockOSThread() // unlock thread
	fmt.Printf("\n\n\n\nCall COBOL XDUMP2C returns %d\n", ret)
}
