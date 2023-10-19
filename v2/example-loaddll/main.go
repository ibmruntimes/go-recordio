package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
//	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func foo() {
	var dll utils.Dll
	var e error
	if len(os.Args) > 1 {
		e = dll.Open(os.Args[1])
	} else {
		e = dll.Open("CDAEQED")
	}
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	defer func() {
		dll.Close()
	}()
	dll.ResolveAll()
	fmt.Printf("The list of functions we found in the DLL\n\n")
	for k, v := range dll.Symbols {
		fmt.Printf("%s -> @%x\n", k, v)
	}
	fmt.Printf("\nRefer to COBOL example for calling functions in a DLL\n")
}
func main() {
	foo()
}
