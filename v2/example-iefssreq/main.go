package main

import (
	"fmt"
	"log"
	"reflect"
	"unsafe"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

type Ssvi struct {
	SSVILEN  uint16
	SSVIVER  byte
	SSVIRSV1 byte
	SSVIID   [4]byte
	SSVIRLEN uint16
	SSVIRVER byte
	SSVIRSV2 byte
	SSVIFLEN uint16
	SSVIASID uint16
	SSVIVERS [8]byte
	SSVIFMID [8]byte
	SSVICNAM [8]byte
	SSVIUDOF uint32
	SSVISDOF uint32
	SSVIPLVL byte
	SSVISLVL byte
	_        [14]byte
	SSVIVLEN uint16
	_        [8]byte // variable length area
}
type Ssob struct {
	SSOBID   [4]byte
	SSOBLEN  uint16
	SSOBFUNC uint16
	SSOBSSIB uint32
	SSOBRETN uint32
	SSOBINDV uint32
	SSOBRETA uint32
	SSOBFLG1 byte
	SSOBRSV1 [3]byte
}
type Ssib struct {
	SSIBID   [4]byte
	SSIBLEN  uint16
	SSIBFLG1 byte
	SSIBSSID byte
	SSIBSSNM [4]byte
	SSIBJBID [8]byte
	SSIBDEST [8]byte
	SSIBRSV1 uint32
	SSIBSUSE uint32
}

type WorkArea32 struct {
	addrSsob0 uint32
	addrSsob1 uint32
	ssob      Ssob
	ssib      Ssib
	ssvi      Ssvi
	dsa       [18]uint32
}

func main() {
	w32 := utils.Malloc31(int((reflect.TypeOf((*WorkArea32)(nil)).Elem()).Size()))
	if uintptr(w32) == 0 {
		log.Fatalf("Failed to malloc31")
	}
	defer utils.Free(w32)
	var w32ptr *WorkArea32
	w32ptr = (*WorkArea32)(w32)
	w32ptr.addrSsob1 = uint32(uintptr(unsafe.Pointer(&(w32ptr.ssob)))) | 0x80000000
	w32ptr.ssob.SSOBID = [4]byte{'S', 'S', 'O', 'B'}
	utils.AtoE(w32ptr.ssob.SSOBID[:])
	w32ptr.ssob.SSOBLEN = uint16((reflect.TypeOf((*Ssob)(nil)).Elem()).Size())
	w32ptr.ssob.SSOBSSIB = uint32(uintptr(unsafe.Pointer(&w32ptr.ssib)))
	w32ptr.ssob.SSOBFUNC = 54 // subsystem verstion information
	w32ptr.ssob.SSOBINDV = uint32(uintptr(unsafe.Pointer(&w32ptr.ssvi)))
	w32ptr.ssib.SSIBID = [4]byte{'S', 'S', 'I', 'B'}
	utils.AtoE(w32ptr.ssib.SSIBID[:])
	w32ptr.ssib.SSIBLEN = uint16((reflect.TypeOf((*Ssib)(nil)).Elem()).Size())
	w32ptr.ssib.SSIBSSNM = [4]byte{'M', 'S', 'T', 'R'}
	utils.AtoE(w32ptr.ssib.SSIBSSNM[:])
	w32ptr.ssvi.SSVILEN = uint16((reflect.TypeOf((*Ssvi)(nil)).Elem()).Size())
	w32ptr.ssvi.SSVIVER = 2
	w32ptr.ssvi.SSVIID = [4]byte{'S', 'S', 'V', 'I'}
	utils.AtoE(w32ptr.ssvi.SSVIID[:])
	ret := utils.Iefssreq(unsafe.Pointer(&w32ptr.addrSsob1), unsafe.Pointer(&w32ptr.dsa[0]))
	if ret == 0 {
		utils.EtoA(w32ptr.ssvi.SSVIVERS[:])
		fmt.Printf("VERS: %s\n", string(w32ptr.ssvi.SSVIVERS[:]))
		utils.EtoA(w32ptr.ssvi.SSVIFMID[:])
		fmt.Printf("FMID: %s\n", string(w32ptr.ssvi.SSVIFMID[:]))
		utils.EtoA(w32ptr.ssvi.SSVICNAM[:])
		fmt.Printf("CNAM: %s\n", string(w32ptr.ssvi.SSVICNAM[:]))
		// debug:  runtime.HexDump(uintptr(w32), (reflect.TypeOf((*WorkArea32)(nil)).Elem()).Size())
	} else {
		log.Fatalf("IEFSSREQ return code %x\n", ret)
	}
}
