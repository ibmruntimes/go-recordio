package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"

	//	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

type DwarfError uintptr
type DwarfDebug uintptr
type DwarfSection uintptr

type Dwarf_Bool int
type Dwarf_Off uint64
type Dwarf_Unsigned uint64
type Dwarf_Half uint16
type Dwarf_Small byte
type Dwarf_Signed int64
type Dwarf_Addr uint64

type Dwarf_Ptr uintptr

type Dwarf_Flag uint32
type Dwarf_Tag uint64

const (
	DW_DLV_OK                = 0
	DW_SECTION_IS_DEBUG_DATA = 0
	DW_SECTION_IS_DEBUG_REL  = 1
	DW_SECTION_IS_DEBUG_RELA = 2
)

func getFunc(dll *utils.Dll, str string) uintptr {
	fp, e := dll.Sym(str)
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	return fp
}

var sectname = [...]string{
	"DW_SECTION_DEBUG_INFO", "DW_SECTION_DEBUG_LINE",
	"DW_SECTION_DEBUG_ABBREV", "DW_SECTION_DEBUG_FRAME",
	"DW_SECTION_EH_FRAME", "DW_SECTION_DEBUG_ARANGES",
	"DW_SECTION_DEBUG_RANGES", "DW_SECTION_DEBUG_PUBNAMES",
	"DW_SECTION_DEBUG_PUBTYPES", "DW_SECTION_DEBUG_STR",
	"DW_SECTION_DEBUG_FUNCNAMES", "DW_SECTION_DEBUG_VARNAMES",
	"DW_SECTION_DEBUG_WEAKNAMES", "DW_SECTION_DEBUG_MACINFO",
	"DW_SECTION_DEBUG_LOC", "DW_SECTION_DEBUG_PPA",
	"DW_SECTION_DEBUG_SRCFILES", "DW_SECTION_DEBUG_SRCTEXT",
	"DW_SECTION_DEBUG_SRCATTR", "DW_SECTION_DEBUG_XREF",
	"DW_SECTION_DEBUG_TYPE"}

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
	fmt.Printf("Function\t\t\t\t\t\t\tAddress\n")
	fmt.Printf("--------\t\t\t\t\t\t\t-------\n")
	for k, v := range dll.Symbols {
		tk := (len(k)) / 8
		if tk > 7 {
			tk = 7
		}
		tabs := 8 - tk
		fmt.Printf("%s%s@%x\n", k, strings.Repeat("\t", tabs), v)
	}
	fmt.Printf("\nRefer to COBOL example for calling functions in a DLL\n\n\n")
	var progname [1025]byte
	var dbg DwarfDebug
	var err DwarfDebug
	copy(progname[:], os.Args[0]+"\x00")
	utils.AtoE(progname[:(len(os.Args[0]) + 1)])
	rc := utils.Cfunc(getFunc(&dll, "dwarf_goff_init_with_PO_filename"), uintptr(unsafe.Pointer(&progname[0])), 0, 0, 0, uintptr(unsafe.Pointer(&dbg)), uintptr(unsafe.Pointer(&err)))
	if rc != 0 {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	defer utils.Cfunc(getFunc(&dll, "dwarf_finish"), uintptr(unsafe.Pointer(dbg)), uintptr(unsafe.Pointer(&err)))
	var section DwarfSection
	for i, name := range sectname {
		rc = utils.Cfunc(getFunc(&dll, "dwarf_debug_section"),
			uintptr(unsafe.Pointer(dbg)),
			uintptr(i),
			DW_SECTION_IS_DEBUG_DATA,
			uintptr(unsafe.Pointer(&section)),
			uintptr(unsafe.Pointer(&err)))
		if rc == 0 {
			fmt.Printf("Found section %s\n", name)
			var unit_header_length Dwarf_Unsigned
			var version_stamp Dwarf_Half
			var unit_offset Dwarf_Off
			var abbrev_offset Dwarf_Off
			var address_size Dwarf_Half
			var next_unit_offset Dwarf_Off
			unit_header_length = 0
			version_stamp = 0
			unit_offset = 0
			abbrev_offset = 0
			address_size = 0
			next_unit_offset = 0
			rc = utils.Cfunc(getFunc(&dll, "dwarf_next_unit_header"),
				uintptr(unsafe.Pointer(dbg)),
				uintptr(unsafe.Pointer(section)),
				uintptr(unsafe.Pointer(&unit_header_length)),
				uintptr(unsafe.Pointer(&version_stamp)),
				uintptr(unsafe.Pointer(&abbrev_offset)),
				uintptr(unsafe.Pointer(&address_size)),
				uintptr(unsafe.Pointer(&next_unit_offset)),
				uintptr(unsafe.Pointer(&err)))
			for rc == DW_DLV_OK {
				fmt.Printf("header_len %v, version %v, offset %v, address size %v, next unit offset %v\n", unit_header_length, version_stamp, unit_offset, address_size, next_unit_offset)
				unit_offset = next_unit_offset
				rc = utils.Cfunc(getFunc(&dll, "dwarf_next_unit_header"),
					uintptr(unsafe.Pointer(dbg)),
					uintptr(unsafe.Pointer(section)),
					uintptr(unsafe.Pointer(&unit_header_length)),
					uintptr(unsafe.Pointer(&version_stamp)),
					uintptr(unsafe.Pointer(&abbrev_offset)),
					uintptr(unsafe.Pointer(&address_size)),
					uintptr(unsafe.Pointer(&next_unit_offset)),
					uintptr(unsafe.Pointer(&err)))
			}

		}
	}

}
func main() {
	foo()
}
