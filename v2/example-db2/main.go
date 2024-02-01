package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

const (
	SQL_BINARY                       = (0xffffffffffffffff - 2 + 1)
	SQL_C_BINARY                     = SQL_BINARY
	SQL_BIGINT                       = -5
	SQL_C_CHAR                       = SQL_CHAR
	SQL_C_LONG                       = SQL_INTEGER
	SQL_C_SHORT                      = SQL_SMALLINT
	SQL_C_FLOAT                      = SQL_REAL
	SQL_C_DOUBLE                     = SQL_DOUBLE
	SQL_C_BIGINT                     = SQL_BIGINT
	SQL_C_NUMERIC                    = SQL_NUMERIC
	SQL_C_DEFAULT                    = 99
	DB2CLI_VER                       = 0x0210
	SQL_SQLSTATE_SIZE                = 5
	SQL_MAX_MESSAGE_LENGTH           = 1024
	SQL_MAX_DSN_LENGTH               = 32
	SQL_MAX_ID_LENGTH                = 18
	SQL_HANDLE_ENV                   = 1
	SQL_HANDLE_DBC                   = 2
	SQL_HANDLE_STMT                  = 3
	SQL_HANDLE_DESC                  = 4
	SQL_NULL_HANDLE                  = 0
	SQL_NORC                         = 0
	SQL_SUCCESS                      = 0
	SQL_SUCCESS_WITH_INFO            = 1
	SQL_NEED_DATA                    = 99
	SQL_NO_DATA                      = 100
	SQL_STILL_EXECUTING              = 2
	SQL_ERROR                        = -1
	SQL_INVALID_HANDLE               = -2
	SQL_CLOSE                        = 0
	SQL_DROP                         = 1
	SQL_UNBIND                       = 2
	SQL_RESET_PARAMS                 = 3
	SQL_COMMIT                       = 0
	SQL_ROLLBACK                     = 1
	SQL_CHAR                         = 1
	SQL_NUMERIC                      = 2
	SQL_DECIMAL                      = 3
	SQL_INTEGER                      = 4
	SQL_SMALLINT                     = 5
	SQL_FLOAT                        = 6
	SQL_REAL                         = 7
	SQL_DOUBLE                       = 8
	SQL_DATE                         = 9
	SQL_TIME                         = 10
	SQL_TIMESTAMP                    = 11
	SQL_VARCHAR                      = 12
	SQL_DATETIME                     = SQL_DATE
	SQL_DECFLOAT                     = (-360)
	SQL_CODE_DATE                    = 1
	SQL_CODE_TIME                    = 2
	SQL_CODE_TIMESTAMP               = 3
	SQL_CODE_TIMESTAMP_WITH_TIMEZONE = 4
	SQL_TYPE_MIN                     = SQL_CHAR
	SQL_TYPE_MAX                     = SQL_VARCHAR
	SQL_TYPE_DATE                    = 91
	SQL_TYPE_TIME                    = 92
	SQL_TYPE_TIMESTAMP               = 93
	SQL_TYPE_TIMESTAMP_WITH_TIMEZONE = 95
	SQL_UNSPECIFIED                  = 0
	SQL_INSENSITIVE                  = 1
	SQL_SENSITIVE                    = 2
	SQL_WCHAR                        = (-8)
	SQL_WVARCHAR                     = (-9)
	SQL_WLONGVARCHAR                 = (-10)
	SQL_GRAPHIC                      = (-95)
	SQL_VARGRAPHIC                   = (-96)
	SQL_LONGVARGRAPHIC               = (-97)
	SQL_BLOB                         = (-98)
	SQL_CLOB                         = (-99)
	SQL_DBCLOB                       = (-350)
	SQL_XML                          = (-370)
	SQL_ROWID                        = (-100)
	SQL_C_DBCHAR                     = SQL_DBCLOB
	SQL_C_ROWID                      = SQL_ROWID
	SQL_C_WCHAR                      = SQL_WCHAR
	SQL_C_DECIMAL64                  = SQL_DECFLOAT
	SQL_C_DECIMAL128                 = -361
	SQL_C_BINARYXML                  = -371
	SQL_BLOB_LOCATOR                 = 31
	SQL_CLOB_LOCATOR                 = 41
	SQL_DBCLOB_LOCATOR               = -351
	SQL_C_BLOB_LOCATOR               = SQL_BLOB_LOCATOR
	SQL_C_CLOB_LOCATOR               = SQL_CLOB_LOCATOR
	SQL_C_DBCLOB_LOCATOR             = SQL_DBCLOB_LOCATOR
	SQL_NO_NULLS                     = 0
	SQL_NULLABLE                     = 1
	SQL_NULLABLE_UNKNOWN             = 2
	SQL_NAMED                        = 0
	SQL_UNNAMED                      = 1
	SQL_PRED_NONE                    = 0
	SQL_PRED_CHAR                    = 1
	SQL_PRED_BASIC                   = 2
	SQL_NULL_DATA                    = -1
	SQL_DATA_AT_EXEC                 = -2
	//SQL_NTS                          = -3
	SQL_NTS                         = (0xffffffffffffffff - 3 + 1)
	SQL_COLUMN_COUNT                = 0
	SQL_COLUMN_NAME                 = 1
	SQL_COLUMN_TYPE                 = 2
	SQL_COLUMN_LENGTH               = 3
	SQL_COLUMN_PRECISION            = 4
	SQL_COLUMN_SCALE                = 5
	SQL_COLUMN_DISPLAY_SIZE         = 6
	SQL_COLUMN_NULLABLE             = 7
	SQL_COLUMN_UNSIGNED             = 8
	SQL_COLUMN_MONEY                = 9
	SQL_COLUMN_UPDATABLE            = 10
	SQL_COLUMN_AUTO_INCREMENT       = 11
	SQL_COLUMN_CASE_SENSITIVE       = 12
	SQL_COLUMN_SEARCHABLE           = 13
	SQL_COLUMN_TYPE_NAME            = 14
	SQL_COLUMN_TABLE_NAME           = 15
	SQL_COLUMN_OWNER_NAME           = 16
	SQL_COLUMN_QUALIFIER_NAME       = 17
	SQL_COLUMN_LABEL                = 18
	SQL_COLUMN_SCHEMA_NAME          = SQL_COLUMN_OWNER_NAME
	SQL_COLUMN_CATALOG_NAME         = SQL_COLUMN_QUALIFIER_NAME
	SQL_COLUMN_DISTINCT_TYPE        = 1250
	SQL_COLUMN_REFERENCE_TYPE       = 1251
	SQL_DESC_DISTINCT_TYPE          = SQL_COLUMN_DISTINCT_TYPE
	SQL_DESC_REFERENCE_TYPE         = SQL_COLUMN_REFERENCE_TYPE
	SQL_DESC_COUNT                  = 1001
	SQL_DESC_TYPE                   = 1002
	SQL_DESC_LENGTH                 = 1003
	SQL_DESC_OCTET_LENGTH_PTR       = 1004
	SQL_DESC_PRECISION              = 1005
	SQL_DESC_SCALE                  = 1006
	SQL_DESC_DATETIME_INTERVAL_CODE = 1007
	SQL_DESC_NULLABLE               = 1008
	SQL_DESC_INDICATOR_PTR          = 1009
	SQL_DESC_DATA_PTR               = 1010
	SQL_DESC_NAME                   = 1011
	SQL_DESC_UNNAMED                = 1012
	SQL_DESC_OCTET_LENGTH           = 1013
	SQL_DESC_ALLOC_TYPE             = 1099
	SQL_NULL_HENV                   = 0
	SQL_NULL_HDBC                   = 0
	SQL_NULL_HSTMT                  = 0
	SQL_DECIMAL64_LEN               = 8
	SQL_DECIMAL128_LEN              = 16
)

func getFunc(dll *utils.Dll, str string) uintptr {
	fp, e := dll.Sym(str)
	if e != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Fatalf("[FATAL] %s [%s:%s:%d]", e, runtime.FuncForPC(pc).Name(), fn, line)
	}
	return fp
}

func foo() {
	var dll utils.Dll
	var e error
	if len(os.Args) > 1 {
		e = dll.Open(os.Args[1])
	} else {
		e = dll.Open("DSNAO64C")
	}
	if e != nil {
		fmt.Fprintf(os.Stderr, "Failed to load DLL %s\n", "DSNAO64C")
		os.Exit(1)
	}
	defer func() {
		dll.Close()
	}()
	fmt.Printf("Load successful\n")

	var henv int32

	SQLAllocHandle := getFunc(&dll, "SQLAllocHandle")
	SQLFreeHandle := getFunc(&dll, "SQLFreeHandle")
	rc := int16(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_ENV, 0, uintptr(unsafe.Pointer(&henv))))
	fmt.Printf("SQLAllocHandle SQL_HANDLE_ENV,0,&henv:  rc %d henv %+v\n", rc, henv)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_ENV, uintptr(henv))
	} else {
		return
	}

	var hdbc int32
	rc = int16(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_DBC, uintptr(henv), uintptr(unsafe.Pointer(&hdbc))))
	fmt.Printf("SQLAllocHandle SQL_HANDLE_DBC,henv,&hdbc:  rc %d hdbc %+v\n", rc, hdbc)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_DBC, uintptr(hdbc))
	} else {
		return
	}

	rc = int16(utils.CfuncEbcdic(getFunc(&dll, "SQLConnect"), uintptr(hdbc), 0, 0, 0, 0, 0, 0))
	fmt.Printf("SQLConnect rc %d\n", rc)

	var hstmt int32
	rc = int16(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_STMT, uintptr(hdbc), uintptr(unsafe.Pointer(&hstmt))))
	fmt.Printf("SQLAllocHandle SQL_HANDLE_STMT,hdbc,&hstmt:  rc %d hstmt %+v\n", rc, hstmt)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_STMT, uintptr(hstmt))
	} else {
		return
	}
	var sqlstmt = []byte("SELECT * FROM DSN81210.EMP")
	sqlstmtsz := uintptr(len(sqlstmt))
	rc = int16(utils.CfuncEbcdic(getFunc(&dll, "SQLExecDirect"), uintptr(hstmt), uintptr(unsafe.Pointer(&sqlstmt[0])), sqlstmtsz))
	fmt.Printf("SQLExecDirect sqlstmt:  rc %d\n", rc)
	if rc != SQL_SUCCESS && rc != SQL_SUCCESS_WITH_INFO {
		return
	}

	var cols int16
	rc = int16(utils.CfuncEbcdic(getFunc(&dll, "SQLNumResultCols"), uintptr(hstmt), uintptr(unsafe.Pointer(&cols))))
	fmt.Printf("SQLNumResultCols:  rc %d cols %d\n", rc, cols)
	if rc != SQL_SUCCESS && rc != SQL_SUCCESS_WITH_INFO {
		return
	}

	SQLDescribeCol := getFunc(&dll, "SQLDescribeCol")
	SQLBindCol := getFunc(&dll, "SQLBindCol")
	type column_t struct {
		colname     [1024]byte
		pcbcolname  int16
		coltype     int16
		coldef      uint32
		scale       uint16
		nullable    uint16
		buffer      []byte
		bytesreturn uint32
	}
	c := make([]column_t, cols, cols)

	for i := 0; i < int(cols); i++ {
		rc = int16(utils.CfuncEbcdic(SQLDescribeCol, uintptr(hstmt), uintptr(i+1),
			uintptr(unsafe.Pointer(&c[i].colname[0])), uintptr(len(c[i].colname)),
			uintptr(unsafe.Pointer(&c[i].pcbcolname)),
			uintptr(unsafe.Pointer(&c[i].coltype)),
			uintptr(unsafe.Pointer(&c[i].coldef)),
			uintptr(unsafe.Pointer(&c[i].scale)),
			uintptr(unsafe.Pointer(&c[i].nullable))))
		if rc != SQL_SUCCESS {
			return
		}
		fmt.Printf("column %d label %s type %d\n", i+1, string(c[i].colname[:c[i].pcbcolname]), c[i].coltype)
		c[i].buffer = make([]byte, (1000 + c[i].coldef))
		rc = int16(utils.CfuncEbcdic(SQLBindCol, uintptr(hstmt), uintptr(i+1),
			uintptr(c[i].coltype),
			uintptr(unsafe.Pointer(&c[i].buffer[0])),
			uintptr(1000+c[i].coldef),
			uintptr(unsafe.Pointer(&c[i].bytesreturn)),
		))
		if rc == -1 {
			rc = int16(utils.CfuncEbcdic(SQLBindCol, uintptr(hstmt), uintptr(i+1),
				uintptr(SQL_BINARY),
				uintptr(unsafe.Pointer(&c[i].buffer[0])),
				uintptr(1000+c[i].coldef),
				uintptr(unsafe.Pointer(&c[i].bytesreturn)),
			))
		}

	}
	if rc != SQL_SUCCESS && rc != SQL_SUCCESS_WITH_INFO {
		return
	}

}
func main() {
	foo()
}
