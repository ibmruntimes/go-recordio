//go:build zos
// +build zos

package utils

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

var Trace bool

func Clib(offset uintptr, parms ...uintptr) (ret uintptr) {
	ret = runtime.CallLeFuncByPtr(runtime.XplinkLibvec+offset<<4, parms)
	return
}
func Cfunc(funcptr uintptr, parms ...uintptr) (ret uintptr) {
	ret = runtime.CallLeFuncByPtr(funcptr, parms)
	return
}
func CfuncEbcdic(funcptr uintptr, parms ...uintptr) (ret uintptr) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	ThreadEbcdicMode()
	defer ThreadAsciiMode()
	ret = runtime.CallLeFuncByPtr(funcptr, parms)
	return
}
func ToCStringPointer(str string) uintptr {
	return uintptr(unsafe.Pointer(&[]byte(str + "\x00")[0]))
}
func ThreadAsciiMode() {
	Clib(0x791, 1)
}
func ThreadEbcdicMode() {
	Clib(0x791, 0)
}
func Malloc64(size int) (ret unsafe.Pointer) {
	ret = unsafe.Pointer(runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_MALLOC<<4,
		[]uintptr{uintptr(size)}))
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_MEMSET<<4, []uintptr{uintptr(ret), 0, uintptr(size)})
	return
}
func Malloc31(size int) (ret unsafe.Pointer) {
	ret = unsafe.Pointer(runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___MALLOC31<<4,
		[]uintptr{uintptr(size)}))
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_MEMSET<<4, []uintptr{uintptr(ret), 0, uintptr(size)})
	return
}
func Malloc24(size int) (ret unsafe.Pointer) {
	ret = unsafe.Pointer(runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___MALLOC24<<4,
		[]uintptr{uintptr(size)}))
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_MEMSET<<4, []uintptr{uintptr(ret), 0, uintptr(size)})
	return
}
func Free(ptr unsafe.Pointer) {
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_FREE<<4,
		[]uintptr{uintptr(ptr)})
}

func _etoA(record []byte) {
	sz := len(record)
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___E2A_L<<4, // __e2a_l
		[]uintptr{uintptr(unsafe.Pointer(&record[0])), uintptr(sz)})
}

func _atoE(record []byte) {
	sz := len(record)
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___A2E_L<<4, // __a2e_l
		[]uintptr{uintptr(unsafe.Pointer(&record[0])), uintptr(sz)})
}

var EtoA = _etoA
var AtoE = _atoE

//go:noescape
func Bpxcall(plist []unsafe.Pointer, bpx_offset int64)

//go:nosplit
func IefssreqX(parm unsafe.Pointer, branch_ptr unsafe.Pointer, save_area unsafe.Pointer) uintptr

//go:nosplit
func Pc31(pcno uintptr, parm unsafe.Pointer) (rc uintptr, rn uintptr)

//go:nosplit
func Svc8(r0 unsafe.Pointer, r1 uintptr) (rr0 unsafe.Pointer, rr1 uintptr, rr15 uintptr)

//go:nosplit
func Svc9(EntryPointName unsafe.Pointer) (r15 uintptr)

//go:nosplit
func Call24(p *ModuleInfo) uintptr

//go:nosplit
func Call31(p *ModuleInfo) uintptr

//go:nosplit
func Call64(p *ModuleInfo) uintptr

//go:nosplit
func Deref(ptr uintptr) (value uintptr, error uintptr)

func Iefssreq(parm unsafe.Pointer, dsa unsafe.Pointer) (ret uintptr) {
	branch_ptr := unsafe.Pointer(uintptr(*(*int32)(unsafe.Pointer(uintptr(*(*uint32)(unsafe.Pointer(uintptr(*(*uint32)(unsafe.Pointer(uintptr(0) + 16))) + 296))) + 20))))
	ret = IefssreqX(parm, branch_ptr, dsa)
	return
}

const (
	Amode24 = 1
	Amode31 = 2
	Amode64 = 3
)

type Thunk24 struct {
	Sam24   uint16
	Basr    uint16
	Sam64   uint16
	Loadr14 [3]uint16
	Br14    uint16
	_       uint16
	Braddr  uintptr
}

func (p *Thunk24) Init() {
	p.Sam24 = 0x010c      // sam24
	p.Basr = 0x0def       // basr 14,15
	p.Sam64 = 0x010e      // sam6
	p.Loadr14[0] = 0xc4e8 // lgrl 14,+8
	p.Loadr14[1] = 0x0000
	p.Loadr14[2] = uint16((uintptr(unsafe.Pointer(&p.Braddr)) - uintptr(unsafe.Pointer(&p.Loadr14[0]))) / 2)
	p.Br14 = 0x07fe

}
func (p *ModuleInfo) Free() {
	if Trace {
		fmt.Printf("Free(%v) ", *p)
	}
	rc := Svc9(unsafe.Pointer(&p.Modname[0]))
	if Trace {
		fmt.Printf(" rc %x \n", 0xffffffff&rc)
	}
	Free(unsafe.Pointer(p))
}

type ModuleInfo struct {
	Modname [8]byte
	Entry   uintptr
	R1      uintptr
	R13     unsafe.Pointer
	R15     uintptr
	Amode   uintptr
	Thunk   Thunk24
	DSA     [144]byte
}

func LoadMod(name string) (ret *ModuleInfo) {
	if Trace {
		fmt.Printf("LOAD %s\n", name)
	}
	var p *ModuleInfo
	p = (*ModuleInfo)(Malloc24(int((reflect.TypeOf((*ModuleInfo)(nil)).Elem()).Size())))
	if uintptr(unsafe.Pointer(p)) == 0 {
		if Trace {
			fmt.Printf("Malloc24 failed\n")
		}
		return
	}
	copy(p.Modname[:], name)
	if len(name) < 8 {
		copy(p.Modname[len(name):], "        ")
	}
	AtoE(p.Modname[:])
	var r0 unsafe.Pointer
	var r1, r15 uintptr
	r1 = 0x0000000080000000
	r0 = unsafe.Pointer(&(p.Modname[0]))
	r0, r1, r15 = Svc8(r0, r1)
	p.R15 = 0x00000000ffffffff & r15 // only lower 31 bit is meaningful
	if p.R15 == 0 {
		p.R1 = r1
		p.R13 = unsafe.Pointer(&p.DSA[0])
		if (0x01 & uintptr(r0)) == 0x01 {
			if Trace {
				fmt.Printf("AMODE 64\n")
			}
			p.Amode = Amode64
			p.Entry = uintptr(r0) & 0xfffffffffffffffe
		} else if (0x0000000080000000 & uintptr(r0)) == 0x0000000080000000 {
			if Trace {
				fmt.Printf("AMODE 31\n")
			}
			p.Amode = Amode31
			p.Entry = uintptr(r0) & 0x000000007fffffff
		} else {
			if Trace {
				fmt.Printf("AMODE 24\n")
			}
			p.Amode = Amode24
			p.Entry = uintptr(r0) & 0x000000007fffffff
			p.Thunk.Init()
		}
		ret = p
	} else {
		if Trace {
			fmt.Printf("svc 8 failed R15=%x\n", p.R15)
		}
		Free(unsafe.Pointer(p))
	}
	return
}

func (p *ModuleInfo) Call(plist uintptr) (ret uintptr) {
	if Trace {
		fmt.Printf("Call %x %x\n", p, plist)
	}
	p.R1 = plist
	if Trace {
		begin := uintptr(unsafe.Pointer(&p.Modname[0]))
		fmt.Printf("Offsets__________\n")
		fmt.Printf("Modename     %d\n", 0)
		fmt.Printf("Entry        %d\n", uintptr(unsafe.Pointer(&p.Entry))-begin)
		fmt.Printf("R1           %d\n", uintptr(unsafe.Pointer(&p.R1))-begin)
		fmt.Printf("R13          %d\n", uintptr(unsafe.Pointer(&p.R13))-begin)
		fmt.Printf("R15          %d\n", uintptr(unsafe.Pointer(&p.R15))-begin)
		fmt.Printf("Amode        %d\n", uintptr(unsafe.Pointer(&p.Amode))-begin)
		fmt.Printf("Thunk.Sam24  %d\n", uintptr(unsafe.Pointer(&p.Thunk.Sam24))-begin)
		fmt.Printf("Thunk.Br14   %d\n", uintptr(unsafe.Pointer(&p.Thunk.Br14))-begin)
		fmt.Printf("Thunk.Braddr %d\n", uintptr(unsafe.Pointer(&p.Thunk.Braddr))-begin)
		fmt.Printf("DSA          %d\n", uintptr(unsafe.Pointer(&p.DSA[0]))-begin)
		runtime.HexDump(uintptr(unsafe.Pointer(p)), (reflect.TypeOf((*ModuleInfo)(nil)).Elem()).Size())
	}
	if p.Amode == Amode24 {
		if Trace {
			fmt.Printf("Call24\n")
		}
		ret = Call24(p)
	} else if p.Amode == Amode31 {
		if Trace {
			fmt.Printf("Call31\n")
		}
		ret = Call31(p)
	} else if p.Amode == Amode64 {
		if Trace {
			fmt.Printf("Call64\n")
		}
		ret = Call64(p)
	} else {
		if Trace {
			fmt.Printf("Unknown AMODE\n")
			p.R15 = 0xffffffffffffffff
		}
		ret = 0xffffffffffffffff
	}
	if Trace {
		fmt.Printf("return %x\n", ret)
		runtime.HexDump(uintptr(unsafe.Pointer(p)), (reflect.TypeOf((*ModuleInfo)(nil)).Elem()).Size())
	}
	return
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
func ConvertSliceToTypes[T1, T2 any](bi []byte) (*T1, *T2) {
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
func Perror() {
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___PERROR_A<<4, //perror()
		[]uintptr{})
}

// The equivalent of the dup2() function which duplicates a file descriptor
func Dup2(oldfd uintptr, newfd uintptr) uintptr {
	ret := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DUP2<<4,
		[]uintptr{oldfd, newfd})
	return ret
}

type Dll struct {
	Dllname  string
	Handle   uintptr
	Count    uint32
	TablePtr uintptr
	Symbols  map[string]uintptr
	Valid    bool
}

func LeFuncName(funcptr uintptr) (name string) {
	f, e := Deref(8 + funcptr)
	if e != 0 {
		return
	}
	h, e := Deref(f - 16)
	if e != 0 {
		return
	}
	if h != 0xc300c500c500f1 {
		return
	}
	ppaoff, e := Deref(f - 8)
	if e != 0 {
		return
	}
	ppaoff >>= 32
	ppa1 := f - 16 + ppaoff
	flags, e := Deref(ppa1 + 8)
	if e != 0 {
		return
	}
	flags >>= 32
	if 0 == (flags & 0x00000001) {
		// no name
		return
	}
	var optflds uintptr
	if 0 != (flags & 0x00008000) {
		// state variable locator
		optflds++
	}
	if 0 != (flags & 0x00004000) {
		// arg area len
		optflds++
	}
	if 0 != (flags & 0x00003000) {
		// fpr mask or ar mask
		optflds++
	}
	if 0 != (flags & 0x00000800) {
		// member word
		optflds++
	}
	if 0 != (flags & 0x00000400) {
		// ppa3 present
		optflds++
	}
	optflds <<= 2
	namesz1, e := Deref(ppa1 + 0x14 + optflds)
	if e != 0 {
		return
	}
	namesz1 >>= 48
	if namesz1 > 128 {
		return
	}
	var funcname [128]byte
	for i := 0; i < int(namesz1); i += 8 {
		v, e := Deref(ppa1 + 0x16 + optflds + uintptr(i))
		if e != 0 {
			return
		}
		funcname[i] = byte(v >> 56)
		funcname[i+1] = byte(v >> 48)
		funcname[i+2] = byte(v >> 40)
		funcname[i+3] = byte(v >> 32)
		funcname[i+4] = byte(v >> 24)
		funcname[i+5] = byte(v >> 16)
		funcname[i+6] = byte(v >> 8)
		funcname[i+7] = byte(v)
	}
	runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___E2A_L<<4, // __e2a_l
		[]uintptr{uintptr(unsafe.Pointer(&funcname[0])), namesz1})
	name = string(funcname[:namesz1])
	return
}

func (d *Dll) Open(name string) error {
	d.Dllname = name
	strptr := uintptr(unsafe.Pointer(&(([]byte(name + "\x00"))[0])))
	d.Handle = runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___DLOPEN_A<<4, []uintptr{strptr, uintptr(0x00010002)}) // dlopen_a
	if d.Handle == 0 {
		return fmt.Errorf("failed to open dll(1) %s, %s", name, d.Error())
	}
	p1, e := Deref(d.Handle + 8)
	if e != 0 {
		runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		return fmt.Errorf("dll handle not valid(1) %s", name)
	}
	l, e := Deref(p1 + 152)
	if e != 0 {
		runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		return fmt.Errorf("dll handle not valid(2) %s", name)
	}
	if l != 0xc4d3c3c2000000a0 {
		runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		return fmt.Errorf("missing dlcb marker %s", name)
	}
	l, e = Deref(p1 + 0x90)
	if e != 0 {
		runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		return fmt.Errorf("dll handle not valid(3) %s", name)
	}
	d.Count = uint32(l >> 32)
	d.TablePtr, e = Deref(p1 + 0x78)
	if e != 0 {
		runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		return fmt.Errorf("dll handle not valid(4) %s", name)
	}
	d.Symbols = make(map[string]uintptr)
	d.Valid = true
	return nil
}
func (d *Dll) Close() error {
	var res uintptr
	if d.Valid {
		res = runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS_DLCLOSE<<4, []uintptr{d.Handle}) // dlclose
		d.Valid = false
	}
	if res != 0 {
		return fmt.Errorf("dll %s dlclose error: %s", d.Dllname, d.Error())
	}
	return nil
}
func (d *Dll) Error() string {
	str := runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___DLERROR_A<<4, []uintptr{d.Handle}) // dlerror
	ba := (*[1<<30 - 1]byte)(unsafe.Pointer(str))
	size := bytes.IndexByte(ba[:], 0)
	return string(ba[:size:size])
}
func (d *Dll) Sym(fn string) (fptr uintptr, err error) {
	fptr, ok := d.Symbols[fn]
	if !ok {
		fptr = runtime.CallLeFuncByPtr(runtime.XplinkLibvec+SYS___DLSYM_A<<4, []uintptr{d.Handle, uintptr(unsafe.Pointer(&(([]byte(fn + "\x00"))[0])))}) // dlsym
		if fptr == 0 {
			err = fmt.Errorf("Symbol %s not found: %s\n", fn, d.Error())
		} else {
			d.Symbols[fn] = fptr
		}
	}
	return
}
func (d *Dll) ResolveAll() error {
	if !d.Valid {
		return fmt.Errorf("DLL info not valid %s", d.Dllname)
	}
	entry := d.TablePtr
	var i uint32
	for i = 0; i < d.Count; i++ {
		fn := LeFuncName(entry)
		if len(fn) != 0 {
			d.Symbols[fn] = entry
		}
		entry += 16
	}

	return nil
}
