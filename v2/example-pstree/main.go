package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"unsafe"
	"github.com/ibmruntimes/go-recordio/v2/utils"
)

const (
	BPX4GTH           = 1056
	PGTHA_FIRST       = 0
	PGTHA_CURRENT     = 1
	PGTHA_NEXT        = 2
	PGTHA_LAST        = 3
	PGTHA_PROCESS     = 0x80
	PGTHA_CONTTY      = 0x40
	PGTHA_PATH        = 0x20
	PGTHA_COMMAND     = 0x10
	PGTHA_FILEDATA    = 0x08
	PGTHA_THREAD      = 0x04
	PGTHA_PTAG        = 0x02
	PGTHA_COMMANDLONG = 0x01
)

type Pgtha struct {
	Pid        int32
	Thid       [8]byte
	Accesspid  byte
	Accesstid  byte
	Accessasid uint16
	Loginname  [8]byte
	Flag1      byte
	Flag1b2    byte
}
type Pgthb struct {
	Id             uint32
	Pid            int32
	Thid           [8]byte
	Accesspid      byte
	Accesstid      byte
	_              [2]byte
	Len_used       int32
	OffsetProcess  int32
	OffsetContty   int32
	OffsetPath     int32
	OffsetCommand  int32
	OffsetFiledata int32
	OffsetThread   int32
}
type Pgthc struct {
	Id           uint32
	Flag1        byte
	Flag2        byte
	Flag3        byte
	_            byte
	Pid          int32
	Ppid         int32
	Pgpid        int32
	Sid          int32
	Fgpid        int32
	Euid         int32
	Ruid         int32
	Suid         int32
	Egid         int32
	Rgid         int32
	Sgid         int32
	Tsize        int32
	SyscallCount int32
	Usertime     int32
	Systime      int32
	Starttime    int32
	Cntoe        uint16
	Cntptcreated uint16
	Cntthreads   uint16
	Asid         uint16
	Jobname      [8]byte
	Loginname    [8]byte
	Memlimit     int32
}
type Pgthf struct {
	Id      uint32
	Len     uint16
	Command [4096]byte
}

type Outline struct {
	Pid     int32
	Ppid    int32
	Command string
}
type Process struct {
	Pid      int32
	Ppid     int32
	Cmd      string
	Children []int32
}

type PidMap map[int32]*Process

func process(outlines []Outline) {
	pmap := make(PidMap)
	pmap[1] = &Process{1, 0, "--root--", nil}
	for _, value := range outlines {
		p := new(Process)
		pmap[value.Pid] = p
		p.Pid = value.Pid
		p.Ppid = value.Ppid
		p.Cmd = value.Command
	}
	for pid, stuff := range pmap {
		if pid == 1 {
			continue
		}
		if pmap[stuff.Ppid] == nil {
			pmap[1].Children = append(pmap[1].Children, pid)
		} else {
			pmap[stuff.Ppid].Children = append(pmap[stuff.Ppid].Children, pid)
		}
	}
	boxes := make([]byte, 2048)
	for i := range boxes {
		boxes[i] = ' '
	}

	res := doPrint(pmap, 1, 0, boxes)
	fmt.Print(res)
}

var asciiBox bool

func boxString(in []byte) (out string) {
	if asciiBox {
		out = string(in)
	} else {
		var b bytes.Buffer
		sz := len(in)
		for i, v := range in {
			if v == '-' {
				b.WriteRune('\u2500')
			} else if v == '+' {
				b.WriteRune('\u2514')
			} else if v == '|' {
				if i < (sz - 1) {
					if in[i+1] == ' ' {
						b.WriteRune('\u2502')
					} else {
						b.WriteRune('\u251c')
					}
				}
			} else {
				b.WriteByte(v)
			}
		}
		out = b.String()
	}
	return
}

func doPrint(pidmap PidMap, pid int32, depth int, boxes []byte) string {
	buffer := new(bytes.Buffer)
	if pid != 1 {
		fmt.Fprintf(buffer, "%s(%v) %v\n", boxString(boxes[:depth*2]), pid, pidmap[pid].Cmd)
		if boxes[depth*2-2] != '|' {
			boxes[depth*2-2] = ' '
		}
		boxes[depth*2-1] = ' '
	}
	z := len(pidmap[pid].Children)
	for i, child := range pidmap[pid].Children {
		if i == (z - 1) {
			boxes[depth*2] = '+'
			boxes[depth*2+1] = '-'
		} else {
			boxes[depth*2] = '|'
			boxes[depth*2+1] = '-'
		}
		fmt.Fprintf(buffer, doPrint(pidmap, child, depth+1, boxes))
	}
	return buffer.String()
}
func convCommand(in []byte) {
	e2at := [...]byte{

		0x20, 0x01, 0x02, 0x03, 0x9c, 0x09, 0x86, 0x7f, 0x97, 0x8d, 0x8e, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x9d, 0x0a, 0x08, 0x87,
		0x18, 0x19, 0x92, 0x8f, 0x1c, 0x1d, 0x1e, 0x1f, 0x80, 0x81, 0x82, 0x83,
		0x84, 0x85, 0x17, 0x1b, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x05, 0x06, 0x07,
		0x90, 0x91, 0x16, 0x93, 0x94, 0x95, 0x96, 0x04, 0x98, 0x99, 0x9a, 0x9b,
		0x14, 0x15, 0x9e, 0x1a, 0x20, 0xa0, 0xe2, 0xe4, 0xe0, 0xe1, 0xe3, 0xe5,
		0xe7, 0xf1, 0xa2, 0x2e, 0x3c, 0x28, 0x2b, 0x7c, 0x26, 0xe9, 0xea, 0xeb,
		0xe8, 0xed, 0xee, 0xef, 0xec, 0xdf, 0x21, 0x24, 0x2a, 0x29, 0x3b, 0x5e,
		0x2d, 0x2f, 0xc2, 0xc4, 0xc0, 0xc1, 0xc3, 0xc5, 0xc7, 0xd1, 0xa6, 0x2c,
		0x25, 0x5f, 0x3e, 0x3f, 0xf8, 0xc9, 0xca, 0xcb, 0xc8, 0xcd, 0xce, 0xcf,
		0xcc, 0x60, 0x3a, 0x23, 0x40, 0x27, 0x3d, 0x22, 0xd8, 0x61, 0x62, 0x63,
		0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0xab, 0xbb, 0xf0, 0xfd, 0xfe, 0xb1,
		0xb0, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70, 0x71, 0x72, 0xaa, 0xba,
		0xe6, 0xb8, 0xc6, 0xa4, 0xb5, 0x7e, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78,
		0x79, 0x7a, 0xa1, 0xbf, 0xd0, 0x5b, 0xde, 0xae, 0xac, 0xa3, 0xa5, 0xb7,
		0xa9, 0xa7, 0xb6, 0xbc, 0xbd, 0xbe, 0xdd, 0xa8, 0xaf, 0x5d, 0xb4, 0xd7,
		0x7b, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0xad, 0xf4,
		0xf6, 0xf2, 0xf3, 0xf5, 0x7d, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50,
		0x51, 0x52, 0xb9, 0xfb, 0xfc, 0xf9, 0xfa, 0xff, 0x5c, 0xf7, 0x53, 0x54,
		0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0xb2, 0xd4, 0xd6, 0xd2, 0xd3, 0xd5,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0xb3, 0xdb,
		0xdc, 0xd9, 0xda, 0x9f,
	}
	for i := 0; i < len(in); i++ {
		in[i] = e2at[in[i]]
	}
}

func main() {
	flag.BoolVar(&asciiBox, "A", false, "Use ASCII box characters")
	flag.Parse()
	var indata Pgtha
	var indata_len int32
	var outdata [5000]byte
	var outdata_len int32
	var rv int32
	var rc int32
	var rn int32

	indata.Pid = 0
	indata.Accesspid = PGTHA_FIRST
	indata.Flag1 |= (PGTHA_COMMANDLONG | PGTHA_PROCESS)
	outdata_len = 5000
	indata_len = 26

	var parms [7]unsafe.Pointer
	var indata_adr unsafe.Pointer
	var outdata_adr unsafe.Pointer
	indata_adr = unsafe.Pointer(&indata)
	outdata_adr = unsafe.Pointer(&outdata[0])
	parms[0] = unsafe.Pointer(&indata_len)
	parms[1] = unsafe.Pointer(&indata_adr)
	parms[2] = unsafe.Pointer(&outdata_len)
	parms[3] = unsafe.Pointer(&outdata_adr)
	parms[4] = unsafe.Pointer(&rv)
	parms[5] = unsafe.Pointer(&rc)
	parms[6] = unsafe.Pointer(&rn)
	outlines := make([]Outline, 0)
	for true {
		utils.Bpxcall(parms[:], BPX4GTH)
		if rv == -1 {
			if rc != 143 {
				log.Fatalf("BPX4GTH errno %d reason %x\n", rc, rn)
			} else {
				break
			}
		}
		var gthc *Pgthc
		var gthf *Pgthf
		if rv == 0 {
			for i := 0; i < 5000; i += 4 {
				tag := *(*uint32)(unsafe.Pointer(&outdata[i]))
				switch {
				case tag == 0x87a38883:
					gthc = (*Pgthc)(unsafe.Pointer(&outdata[i]))
				case tag == 0x87a38886:
					gthf = (*Pgthf)(unsafe.Pointer(&outdata[i]))
					break
				}
			}
			if uintptr(unsafe.Pointer(gthf)) != 0 && gthf.Len > 0 {
				convCommand(gthf.Command[:gthf.Len-1])
				outlines = append(outlines, Outline{gthc.Pid, gthc.Ppid, string(gthf.Command[:gthf.Len-1])})
			}
			indata.Accesspid = PGTHA_NEXT
			indata.Pid = gthc.Pid
		} else {
			break
		}
	}
	process(outlines)
}
