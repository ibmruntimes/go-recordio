// example-osvrm extracts the z/OS version/release/modification (VRM) level by
// walking the MVS control blocks in main memory
//
//	PSA+16 (0x10)      -> CVT
//	CVT-40             -> CVTPRODN  (7 chars, product name, e.g. "SP7.0.4")
//	CVT-32             -> CVTMDL.. FMID (7 chars, e.g. "HBB77D0")
//	CVT+140            -> CVTECVT   (-> ECVT, extended CVT)
//	ECVT+496 (16 chars) ECVTPNAM  product name, e.g. "z/OS"
//	ECVT+512 (2 chars)  ECVTPVER  version
//	ECVT+514 (2 chars)  ECVTPREL  release
//	ECVT+516 (2 chars)  ECVTPMOD  modification
//
// All character fields are EBCDIC (IBM-1047).
package main

import (
	"fmt"
	"os"
	"strings"
	"unsafe"
)

const (
	psaCVTPtr = 0x10 // PSA: pointer to CVT

	cvtProdN  = -40 // CVTPRODN, 7 chars
	cvtFMID   = -32 // FMID, 7 chars
	cvtECVT   = 140 // CVTECVT, 4-byte pointer to ECVT
	cvtOSLvl1 = 1265

	ecvtPNam = 496 // ECVTPNAM, 16 chars
	ecvtPVer = 512 // ECVTPVER, 2 chars
	ecvtPRel = 514 // ECVTPREL, 2 chars
	ecvtPMod = 516 // ECVTPMOD, 2 chars

	cvtOSLvl1OS390R1 = 0x02 // CVTOSLV1 bit: OS/390 R1 or later (ECVT fields valid)
)

// peek returns n bytes of storage at the given address.
func peek(addr uintptr, n int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(addr)), n)
}

// peek32 reads a big-endian 4-byte value and strips the AMODE bit, yielding a
// 31-bit address.
func peek32(addr uintptr) uintptr {
	return uintptr(*(*uint32)(unsafe.Pointer(addr)) & 0x7fffffff)
}

// e2a converts the subset of IBM-1047 used by CVT/ECVT product fields to ASCII.
func e2a(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		switch {
		case c >= 0xC1 && c <= 0xC9:
			sb.WriteByte('A' + c - 0xC1)
		case c >= 0xD1 && c <= 0xD9:
			sb.WriteByte('J' + c - 0xD1)
		case c >= 0xE2 && c <= 0xE9:
			sb.WriteByte('S' + c - 0xE2)
		case c >= 0x81 && c <= 0x89:
			sb.WriteByte('a' + c - 0x81)
		case c >= 0x91 && c <= 0x99:
			sb.WriteByte('j' + c - 0x91)
		case c >= 0xA2 && c <= 0xA9:
			sb.WriteByte('s' + c - 0xA2)
		case c >= 0xF0 && c <= 0xF9:
			sb.WriteByte('0' + c - 0xF0)
		case c == 0x40:
			sb.WriteByte(' ')
		case c == 0x4B:
			sb.WriteByte('.')
		case c == 0x60:
			sb.WriteByte('-')
		case c == 0x61:
			sb.WriteByte('/')
		default:
			sb.WriteByte('?')
		}
	}
	return sb.String()
}

// atoi2 parses a 2-digit decimal EBCDIC/ASCII string, returning -1 if invalid.
func atoi2(s string) int {
	if len(s) != 2 || s[0] < '0' || s[0] > '9' || s[1] < '0' || s[1] > '9' {
		return -1
	}
	return int(s[0]-'0')*10 + int(s[1]-'0')
}

func main() {
	cvt := peek32(psaCVTPtr)
	if cvt == 0 {
		fmt.Fprintln(os.Stderr, "CVT pointer is null")
		os.Exit(1)
	}

	prodn := strings.TrimRight(e2a(peek(uintptr(int(cvt)+cvtProdN), 7)), " ")
	fmid := strings.TrimRight(e2a(peek(uintptr(int(cvt)+cvtFMID), 7)), " ")

	if peek(cvt+cvtOSLvl1, 1)[0]&cvtOSLvl1OS390R1 == 0 {
		// Pre OS/390 R1: no ECVT product fields.
		fmt.Printf("The MVS version is %s - FMID %s.\n", prodn, fmid)
		return
	}

	ecvt := peek32(cvt + cvtECVT)
	pnam := strings.TrimRight(e2a(peek(ecvt+ecvtPNam, 16)), " ")
	ver := e2a(peek(ecvt+ecvtPVer, 2))
	rel := e2a(peek(ecvt+ecvtPRel, 2))
	mod := e2a(peek(ecvt+ecvtPMod, 2))

	fmt.Printf("Product name : %s\n", pnam)
	fmt.Printf("VRM          : %s.%s.%s\n", ver, rel, mod)
	if v, r, m := atoi2(ver), atoi2(rel), atoi2(mod); v >= 0 && r >= 0 && m >= 0 {
		fmt.Printf("VRM (short)  : V%dR%dM%d\n", v, r, m)
	}
	fmt.Printf("FMID         : %s (%s)\n", fmid, prodn)
}
