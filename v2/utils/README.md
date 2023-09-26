# Utiliy functions

# Usage

* func Malloc31(size int) (ret unsafe.Pointer)
  Allocate memory below the bar (<2G)

* func Malloc24(size int) (ret unsafe.Pointer)
  Allocate memory below the line (<16M)

* func Free(ptr unsafe.Pointer)
  Free allocated memory

* func EtoA(record []byte)
  Convert byte array from EBCDIC to ASCII

* func AtoE(record []byte)
  Convert byte array from ASCII to EBCDIC

* func Bpxcall(plist []unsafe.Pointer, bpx_offset int64)
  Perform a call to the USS assembly services.

* func Svc8(r0 unsafe.Pointer, r1 uintptr) (rr0 unsafe.Pointer, rr1 uintptr, rr15 uintptr)
  Perform the MVS macro SVC 8 (load module)

* func Svc9(EntryPointName unsafe.Pointer) (r15 uintptr)
  Perform the MVS macro SVC 9 (free module)

* func Call24(p *ModuleInfo) uintptr
  Branch to a loaded amode24 program 

* func Call31(p *ModuleInfo) uintptr
  Branch to a loaded amode31 program 

* func Call64(p *ModuleInfo) uintptr
  Branch to a loaded amode64 program 

* func Iefssreq(parm unsafe.Pointer, dsa unsafe.Pointer) (ret uintptr)
  Perform the MVS macro IEFSSREQ

* func LoadMod(name string) (ret *ModuleInfo)
  Load the named module into memory

* func ConvertStringToSlice(s string, bi []byte) (bo []byte)
  Convert Go string to byte array

* func ConvertTypeToSlice[T any](i T) (slice []byte)
  Convert a generic type to byte array

* func ConvertSliceToType[T1 any](bi []byte) (*T1, int)
  Convert a byte slice to a generic type

* func ConvertSliceToTypes[T1, T2 any](bi []byte) (*T1, *T2)
  Convert a byte slice to a generic types

* func ConvertSliceToTypes3[T1, T2, T3 any](bi []byte) (*T1, *T2, *T3)
  Convert a byte slice to a generic types

* func Perror()
  Wrapper for the perror() function

* func Dup2(oldfd uintptr, newfd uintptr) uintptr
  Wrapper for the dup2() function
