package main

/*
#cgo LDFLAGS: XDDLL.x
#include <stdlib.h>
int XDUMP(const char* str, int len);
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func main() {
	str := "hello world"
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	runtime.LockOSThread()   // lock thread
	utils.ThreadEbcdicMode() // COBOL must run in ebcdic
	ret := C.XDUMP(cstr, C.int(len(str)))
	utils.ThreadAsciiMode()  // restore
	runtime.UnlockOSThread() // unlock thread
	fmt.Printf("Call COBOL XDUMP returns %d\n", ret)
}
