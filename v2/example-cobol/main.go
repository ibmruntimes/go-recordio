package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func foo() {
	var dll utils.Dll
	var e error
	if len(os.Args) > 1 {
		e = dll.Open(os.Args[1])
	} else {
		e = dll.Open("./XDDLL")
	}
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	defer func() {
		dll.Close()
	}()
	dll.ResolveAll()
	fmt.Printf("The list of functions we found in the DLL\n")
	for k, v := range dll.Symbols {
		fmt.Printf("%s -> @%x\n", k, v)
	}
	fn, e := dll.Sym("XDUMP")
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	// data to pass to COBOL
	data := uintptr(unsafe.Pointer(&(([]byte("Hello World"))[0])))

	runtime.LockOSThread() // lock thread
	utils.ThreadEbcdicMode() // COBOL must run in ebcdic
	ret := utils.Cfunc(fn,data, 11) // 11 characters
	utils.ThreadAsciiMode() // restore
	runtime.UnlockOSThread() // unlock thread
	fmt.Printf("Call COBOL XDUMP returns %d\n", ret)
}
func main() {
	foo()
}
