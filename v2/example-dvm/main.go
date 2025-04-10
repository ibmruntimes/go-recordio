package main

import (
	"encoding/binary"
	"fmt"
	"github.com/ibmruntimes/go-recordio/v2/utils"
	"reflect"
	"strings"
	"unsafe"
)

// Define PLIST to match PLI structure with memory alignment
type PLIST struct {
	list         [5]uint32
	sendArea     [400]byte
	recieveArea  [300]byte
	msgBuff      [40]byte
	sqldaArea    SQLDA
	DVCB_Connect DVCB
}

// SQLVar represents a variable in the SQLDA structure
type SQLVar struct {
	SQLTYPE int16
	SQLLEN  int16
	SQLDATA int32
	SQLIND  int32
	SQLNAME_LENGTH  int16
	Data    [30]byte
}

// SQLDA represents the SQLDA structure used for Metadata
type SQLDA struct {
	SQLDAID [8]byte
	SQLDABC int32
	SQLN    int16
	SQLD    int16
	SQLVar  [100]SQLVar // Type SQLVAR Array
}

// Define nested structs for DVCB fields with proper alignment and nested fields
type DVCB_CNID_t struct {
	DVCB_CONNECTION     [12]byte
	DVCB_CONNECTED_SSID [4]byte
}
type DVCB_OPTIONS_t struct {
	DVCB_OPT_RECV_MODE   byte
	DVCB_OPT_AUTO_COMMIT byte
	DVCB_OPT_CLOSE_AFTER byte
	DVCB_OPT_REFORMAT    byte
	DVCB_OPT_SQLDA       byte
	DVCB_OPT_PRSV_ROW    byte
	_                    [2]byte // Padding
}
type DVCB_RETURN_FLAGS_t struct {
	DVCB_ROW_RETURNED     byte
	DVCB_SQLCODE_RETURNED byte
	DVCB_MESSAGE_RETURNED byte
	DVCB_SQLDA_RETURNED   byte
	DVCB_END_OF_DATA      byte
	DVCB_ERROR_RETURNED   byte
	DVCB_PARMS_RETURNED   byte
	DVCB_END_OF_RSET      byte
	_                     [2]byte // Padding
}
type DVCB_FLAGS_t struct {
	DVCB_FLAG1 byte
	DVCB_FLAG2 byte
	DVCB_FLAG3 byte
	DVCB_FLAG4 byte
}

// Main DVCB structure
type DVCB struct {
	DVCB_TAG                  [4]byte
	DVCB_VERSION              int16
	DVCB_RUNTIME_ENV          byte
	_                         byte // Reserved
	DVCB_SSID                 [4]byte
	DVCB_REQUEST_CODE         [4]byte
	DVCB_CNID                 DVCB_CNID_t
	DVCB_SERVER_GROUP         [8]byte
	DVCB_USER_PARM            [8]byte
	DVCB_SQL_CODE             int32
	DVCB_DATA_BUFFER_LENGTH   int32
	DVCB_DATA_RETURNED_LENGTH int32
	DVCB_ROWS_PER_RECV_CALL   int32
	DVCB_ROWS_RETURNED        int32
	DVCB_OPTIONS              DVCB_OPTIONS_t
	DVCB_BLOCKING_TIMEOUT     int32
	DVCB_SEND_LENGTH          int32
	DVCB_RETURN_CODE          int32
	DVCB_DB2_SUBSYSTEM        [4]byte
	DVCB_ROW_LENGTH           int32
	DVCB_SQLDA_LENGTH         int32
	DVCB_MESSAGE_LENGTH       int32
	DVCB_MAP_NAME             [50]byte
	DVCB_RETURN_FLAGS         DVCB_RETURN_FLAGS_t
	DVCB_ROW_LIMIT            int32
	DVCB_USERID               [8]byte
	DVCB_PASSWORD             [8]byte
	DVCB_MPR_CLIENT_ID        int16
	DVCB_MPR_NUM_CLIENT       int16
	DVCB_MPR_TIME_OUT         int16
	_                         [58]byte // Padding
	DVCB_FLAGS                DVCB_FLAGS_t
	DVCB_EYECATCH2            [4]byte
	DVCB_ADD_DATA_MODE        int64
	DVCB_ADD_SQLDA            int64
	DVCB_VPD_NUM              int16
	DVCB_VPD_TIME_OUT         int16
	DVCB_VPD_GROUP            [8]byte
	DVCB_VPD_IO               int16
	DVCB_SEG_SIZE             byte
	_                         byte // Padding
	DVCB_PAS_LGTH             int16
	DVCB_PASS                 [100]byte
	_                         [6]byte // Padding
	DVCB_TAG2                 [4]byte
}

func fillDVCB(plist31 *PLIST) {
	plist31.DVCB_Connect = DVCB{
		DVCB_TAG:                [4]byte{'D', 'V', 'C', 'B'},
		DVCB_TAG2:               [4]byte{'D', 'V', 'C', 'B'},
		DVCB_VERSION:            2,
		DVCB_RUNTIME_ENV:        'B', // Batch
		DVCB_SSID:               [4]byte{'A', 'V', 'Z', '1'},
		DVCB_SEND_LENGTH:        int32(len(plist31.sendArea)),
		DVCB_DATA_BUFFER_LENGTH: int32(len(plist31.recieveArea)),
		DVCB_SQLDA_LENGTH:       int32(reflect.TypeOf(plist31.sqldaArea).Size()), // Enable SQLDA - Optional during RECV stage - Ignored in other steps
		DVCB_OPTIONS:            DVCB_OPTIONS_t{},
	}

	// Convert fields to EBCDIC
	utils.AtoE(plist31.DVCB_Connect.DVCB_TAG[:])
	utils.AtoE(plist31.DVCB_Connect.DVCB_TAG2[:])
	utils.AtoE((*[2]byte)(unsafe.Pointer(&plist31.DVCB_Connect.DVCB_VERSION))[:])
	utils.AtoE((*[1]byte)(unsafe.Pointer(&plist31.DVCB_Connect.DVCB_RUNTIME_ENV))[:])
	utils.AtoE(plist31.DVCB_Connect.DVCB_SSID[:])
	utils.AtoE(plist31.sendArea[:])

	// Set up pointers in plist31.list
	plist31.list[0] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.DVCB_Connect)))
	plist31.list[1] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.sendArea)))
	plist31.list[2] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.recieveArea)))
	plist31.list[3] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.msgBuff)))
	plist31.list[4] = uint32(0x0ffffffff & uintptr(unsafe.Pointer(&plist31.sqldaArea)))
	plist31.list[4] |= uint32(0x80000000)
}

var buff = record_T{}

const (
	SQL_INTEGER    = 111 // Unverified value
	SQL_SMALLINT   = 500
	SQL_EBCDIC_STR = 452
)

func parseAndPrintOutput(sqldaArea SQLDA, recA []byte, Length int32, rowsReturned int32) {

	for i := 0; i < int(sqldaArea.SQLD); i++ {
		sqlVar := sqldaArea.SQLVar[i]
		nameLen := int(sqlVar.SQLNAME_LENGTH)

		if nameLen <= 0 || nameLen > len(sqlVar.Data) {
			fmt.Printf("Invalid column name length: %d at column %d\n", nameLen, i)

			continue
		}

		utils.EtoA(sqlVar.Data[:nameLen])
		colName := string(sqlVar.Data[:nameLen])
		fmt.Printf("%s ", colName)
	}
	fmt.Printf("\n")

	for i := 0; i < int(sqldaArea.SQLD); i++ {
		fmt.Printf("%s ", strings.Repeat("=", int(sqldaArea.SQLVar[i].SQLNAME_LENGTH)))
	}
	fmt.Printf("\n")

	for row := 0; row < int(rowsReturned); row++ {
		offset := row * int(Length)
		cursor := offset

		if offset+int(Length) > len(recA) {
			fmt.Printf("Row %d is out of bounds - skipping.\n", row)
			continue
		}

		for col := 0; col < int(sqldaArea.SQLD); col++ {
			colWidth := int(sqldaArea.SQLVar[col].SQLNAME_LENGTH)
			fieldLen := int(sqldaArea.SQLVar[col].SQLLEN)
			SQLTYPE := sqldaArea.SQLVar[col].SQLTYPE

			if cursor+fieldLen > len(recA) {
				fmt.Printf("%-*s ", colWidth, "[ERR]")
				cursor += fieldLen
				continue
			}

			field := recA[cursor : cursor+fieldLen]
			cursor += fieldLen

			switch SQLTYPE {
			case SQL_SMALLINT:
				if len(field) >= 2 {
					val := int(binary.BigEndian.Uint16(field))
					fmt.Printf("%-*d ", colWidth, val)
				} else {
					fmt.Printf("%-*s ", colWidth, "[ERR]")
				}

			case SQL_INTEGER:
				if len(field) >= 4 {
					val := int(binary.BigEndian.Uint32(field))
					fmt.Printf("%-*d ", colWidth, val)
				} else {
					fmt.Printf("%-*s ", colWidth, "[ERR]")
				}

			case SQL_EBCDIC_STR:
				ascii := make([]byte, fieldLen)
				copy(ascii, field)
				utils.EtoA(ascii)
				fmt.Printf("%-*s ", colWidth, strings.TrimSpace(string(ascii)))

			default:
				ascii := make([]byte, fieldLen)
				copy(ascii, field)
				utils.EtoA(ascii)
				fmt.Printf("%-*s ", colWidth, strings.TrimSpace(string(ascii)))
			}
		}
		fmt.Println()
	}

}

type record_T struct {
	keyID     int16
	dataNameL int16
	dataName  [9]byte
	dataDept  int16
	dataJob   [5]byte
	dataYrs   int16
}

func requestDVM(plist31 *PLIST, req string) {
	var code [4]byte
	copy(code[:], req)
	utils.AtoE(code[:])
	copy(plist31.DVCB_Connect.DVCB_REQUEST_CODE[:], code[:])
}

func call(mod2 *utils.ModuleInfo, plist31 *PLIST) {
	if uintptr(unsafe.Pointer(mod2)) != 0 {
		RC := mod2.Call(uintptr(unsafe.Pointer(plist31)))
		if RC != 0 {
			fmt.Printf("RC=0x%x\n", RC)
		}
	} else {
		fmt.Printf("Failed to load module AVZCLIEN\n")
	}
}

func main() {
	// Allocate and initialize PLIST
	siz := (int((reflect.TypeOf((*PLIST)(nil)).Elem()).Size()))
	plist31 := (*PLIST)(unsafe.Pointer(utils.Malloc31(siz)))

	mod2 := utils.LoadMod("AVZCLIEN")
	defer mod2.Free()

	copy(plist31.sqldaArea.SQLDAID[:], "SQLDA   ")
	utils.AtoE(plist31.sqldaArea.SQLDAID[:])

	// SQL Statement
	SQLMessage := `SELECT * FROM DVSQL.STAFFVS LIMIT 10`
	copy(plist31.sendArea[:], SQLMessage)

	// Initialize DVCB with required inital values
	fillDVCB(plist31)

	//// OPEN ////
	plist31.DVCB_Connect.DVCB_OPTIONS.DVCB_OPT_SQLDA = 'Y'
	utils.AtoE((*[1]byte)(unsafe.Pointer(&plist31.DVCB_Connect.DVCB_OPTIONS.DVCB_OPT_SQLDA))[:])
	plist31.sqldaArea.SQLDABC = int32(5*44 + 16) // Allocate space for 5 SQLVAR entries

	requestDVM(plist31, "OPEN")
	call(mod2, plist31)

	//// SEND ////
	requestDVM(plist31, "SEND")
	plist31.DVCB_Connect.DVCB_SEND_LENGTH = int32(len(plist31.sendArea))
	call(mod2, plist31)

	//// RECV ////
	requestDVM(plist31, "RECV")
	call(mod2, plist31)

	// Print data since it contains the following:
	parseAndPrintOutput(
		plist31.sqldaArea,
		plist31.recieveArea[:],
		plist31.DVCB_Connect.DVCB_ROW_LENGTH,
		plist31.DVCB_Connect.DVCB_ROWS_RETURNED,
	)

	//// CLOSE ////
	requestDVM(plist31, "CLOS")
	call(mod2, plist31)

	utils.Free(unsafe.Pointer(plist31))
}
