package main

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

type CsvdynlParm struct {
	Version        byte
	Request        byte
	Flags          byte
	Pos            byte
	LnkLstName     [16]byte
	DsnameAddr     uint32
	DsnameAlet     uint32
	CmdInfoaddr    uint32
	Jobname        [8]byte
	Asid           [2]byte
	_              [2]byte
	ProbDsnameAddr uint32
	ProbDsnameAlet uint32
	AnsAddr        uint32
	AnsAlet        uint32
	AnsLen         uint32
	_              [20]byte
}

const (
	LIST = 6
)

type Dlaahdr struct {
	Num1 uint32
	Num2 uint32
	Len  uint32
	Addr uint32
}

type Dlaals struct {
	NextAddr   uint32
	DlaadsAddr uint32
	Dlaals     uint32
	LnkLstName [16]byte
	Flags      byte
	_          [3]byte
	SeqNum     uint32
	_          uint32
	NumDlaads  uint16
	NumDlaau   uint16
}

type Dlaads struct {
	NextAddr uint32
	Flags    byte
	_        [3]byte
	Volser   [6]byte
	NameLen  uint16
	Dsname   [44]byte
}

type Lnklst struct {
	Volser string
	Dsname string
	Afp    bool
	Sms    bool
}

func main() {
	pc := uintptr(*(*int32)(unsafe.Pointer(uintptr(*(*uint32)(unsafe.Pointer(uintptr(*(*uint32)(unsafe.Pointer(uintptr(0) + 16))) + 193*4))) + 112*4)))
	parm := utils.Malloc31(int((reflect.TypeOf((*CsvdynlParm)(nil)).Elem()).Size()))
	if parm == unsafe.Pointer(uintptr(0)) {
		log.Fatalf("Failed to malloc31 parm")
	}
	defer utils.Free(parm)
	anslen := uint32(2048)
	answer := utils.Malloc31(int(anslen))
	if answer == unsafe.Pointer(uintptr(0)) {
		log.Fatalf("Failed to malloc31 answer")
	}
	defer utils.Free(answer)
	p := (*CsvdynlParm)(parm)
	p.Request = LIST
	p.AnsAddr = uint32(uintptr(answer))
	p.AnsLen = anslen
	copy(p.LnkLstName[:], []byte("CURRENT         "))
	utils.AtoE(p.LnkLstName[:])

	rc, rn := utils.Pc31(pc, parm)
	q := (*Dlaahdr)(answer)
	if rc == 4 && rn == 0x403 {
		anslen = q.Len
		utils.Free(answer)
		answer = utils.Malloc31(int(anslen))
		if answer == unsafe.Pointer(uintptr(0)) {
			log.Fatalf("Failed to malloc31 answer")
		}
		p.AnsAddr = uint32(uintptr(answer))
		p.AnsLen = anslen
		rc, rn = utils.Pc31(pc, parm)
		q = (*Dlaahdr)(answer)
	}
	if rc != 0 {
		log.Fatalf("Failed rc %d reason %d", rc, rn)
	}
	dlaals := (*Dlaals)(unsafe.Pointer(uintptr(q.Addr)))
	utils.EtoA(dlaals.LnkLstName[:])
	fmt.Printf("Name: %s\n", string(dlaals.LnkLstName[:]))
	dlaads := (*Dlaads)(unsafe.Pointer(uintptr(dlaals.DlaadsAddr)))
	lnklst := make([]Lnklst, int(dlaals.NumDlaads))
	for i := 0; i < int(dlaals.NumDlaads); i++ {
		if (dlaads.Flags & 0x80) == 0x80 {
			lnklst[i].Afp = true
		}
		if (dlaads.Flags & 0x20) == 0x20 {
			lnklst[i].Sms = true
		}
		utils.EtoA(dlaads.Volser[:])
		utils.EtoA(dlaads.Dsname[:dlaads.NameLen])
		lnklst[i].Volser = string(dlaads.Volser[:])
		lnklst[i].Dsname = string(dlaads.Dsname[:dlaads.NameLen])
		dlaads = (*Dlaads)(unsafe.Pointer(uintptr(dlaads.NextAddr)))
	}
	for i := 0; i < int(dlaals.NumDlaads); i++ {
		fmt.Printf("%+v\n", lnklst[i])
	}
}
