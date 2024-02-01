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
	MAX_COL_NAME_LEN = 256
	MAX_DATA_LEN     = 1024
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

	var henv SQLHENV

	SQLAllocHandle := getFunc(&dll, "SQLAllocHandle")
	SQLFreeHandle := getFunc(&dll, "SQLFreeHandle")
	rc := SQLRETURN(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_ENV, 0, uintptr(unsafe.Pointer(&henv))))
	showSqlError(&dll, rc, "SQLAllocHandle", SQLHANDLE(henv), SQL_HANDLE_ENV)
	fmt.Printf("SQLAllocHandle SQL_HANDLE_ENV,0,&henv:  rc %d henv %+v\n", rc, henv)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_ENV, uintptr(henv))
	} else {
		return
	}

	var hdbc SQLHDBC
	rc = SQLRETURN(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_DBC, uintptr(henv), uintptr(unsafe.Pointer(&hdbc))))
	showSqlError(&dll, rc, "SQLAllocHandle", SQLHANDLE(henv), SQL_HANDLE_ENV)
	fmt.Printf("SQLAllocHandle SQL_HANDLE_DBC,henv,&hdbc:  rc %d hdbc %+v\n", rc, hdbc)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_DBC, uintptr(hdbc))
	} else {
		return
	}

	rc = SQLRETURN(utils.CfuncEbcdic(getFunc(&dll, "SQLConnect"), uintptr(hdbc), 0, 0, 0, 0, 0, 0))
	showSqlError(&dll, rc, "SQLConnect", SQLHANDLE(hdbc), SQL_HANDLE_DBC)
	fmt.Printf("SQLConnect rc %d\n", rc)

	var hstmt SQLHSTMT
	rc = SQLRETURN(utils.CfuncEbcdic(SQLAllocHandle, SQL_HANDLE_STMT, uintptr(hdbc), uintptr(unsafe.Pointer(&hstmt))))
	showSqlError(&dll, rc, "SQLAllocHandle", SQLHANDLE(hdbc), SQL_HANDLE_DBC)
	fmt.Printf("SQLAllocHandle SQL_HANDLE_STMT,hdbc,&hstmt:  rc %d hstmt %+v\n", rc, hstmt)
	if rc == SQL_SUCCESS || rc == SQL_SUCCESS_WITH_INFO {
		defer utils.CfuncEbcdic(SQLFreeHandle, SQL_HANDLE_STMT, uintptr(hstmt))
	} else {
		return
	}
	var sqlstmt = []byte("SELECT * FROM DSN81210.EMP")
	sqlstmtsz := uintptr(len(sqlstmt))
	rc = SQLRETURN(utils.CfuncEbcdic(getFunc(&dll, "SQLExecDirect"), uintptr(hstmt), uintptr(unsafe.Pointer(&sqlstmt[0])), sqlstmtsz))
	showSqlError(&dll, rc, "SQLExecDirect", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
	fmt.Printf("SQLExecDirect sqlstmt:  rc %d\n", rc)
	if rc != SQL_SUCCESS && rc != SQL_SUCCESS_WITH_INFO {
		return
	}

	var cols SQLSMALLINT
	rc = SQLRETURN(utils.CfuncEbcdic(getFunc(&dll, "SQLNumResultCols"), uintptr(hstmt), uintptr(unsafe.Pointer(&cols))))
	showSqlError(&dll, rc, "SQLNumResultCols", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
	fmt.Printf("SQLNumResultCols:  rc %d cols %d\n", rc, cols)
	if rc != SQL_SUCCESS && rc != SQL_SUCCESS_WITH_INFO {
		return
	}

	SQLDescribeCol := getFunc(&dll, "SQLDescribeCol")
	SQLBindCol := getFunc(&dll, "SQLBindCol")
	SQLFetch := getFunc(&dll, "SQLFetch")
	type column_t struct {
		colname    [MAX_COL_NAME_LEN]byte
		pcbcolname SQLSMALLINT
		coltype    SQLSMALLINT
		coldef     SQLULEN
		scale      SQLSMALLINT
		nullable   SQLSMALLINT
		length     SQLLEN
		data       [MAX_DATA_LEN]byte
	}
	c := make([]column_t, cols, cols)

	for i := 0; i < int(cols); i++ {
		rc = SQLRETURN(utils.CfuncEbcdic(SQLDescribeCol, uintptr(hstmt), uintptr(i+1),
			uintptr(unsafe.Pointer(&c[i].colname[0])), uintptr(len(c[i].colname)),
			uintptr(unsafe.Pointer(&c[i].pcbcolname)),
			uintptr(unsafe.Pointer(&c[i].coltype)),
			uintptr(unsafe.Pointer(&c[i].coldef)),
			uintptr(unsafe.Pointer(&c[i].scale)),
			uintptr(unsafe.Pointer(&c[i].nullable))))
		showSqlError(&dll, rc, "SQLDescribeCol", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
		if rc != SQL_SUCCESS {
			return
		}
		rc1 := SQLRETURN(utils.CfuncEbcdic(SQLBindCol, uintptr(hstmt), uintptr(i+1),
			uintptr(SQL_C_CHAR),
			uintptr(unsafe.Pointer(&c[i].data[0])),
			uintptr(len(c[i].data)),
			uintptr(unsafe.Pointer(&c[i].length))))
		showSqlError(&dll, rc1, "SQLBindCol", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
		if rc1 != SQL_SUCCESS {
			return
		}
	}

	for i := 0; i < int(cols); i++ {
		fmt.Printf("%s\t", string(c[i].colname[:c[i].pcbcolname]))
	}
	fmt.Printf("\n")

	rc = SQLRETURN(utils.CfuncEbcdic(SQLFetch, uintptr(hstmt)))
	showSqlError(&dll, rc, "SQLFetch", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
	for rc != SQL_NO_DATA {
		for i := 0; i < int(cols); i++ {
			fmt.Printf("%s\t", c[i].data)
		}
		fmt.Printf("\n")
		rc = SQLRETURN(utils.CfuncEbcdic(SQLFetch, uintptr(hstmt)))
		showSqlError(&dll, rc, "SQLFetch", SQLHANDLE(hstmt), SQL_HANDLE_STMT)
	}

}

func showSqlError(dll *utils.Dll, er SQLRETURN, message string, h SQLHANDLE, typ SQLSMALLINT) {

	if er != SQL_SUCCESS && er != SQL_SUCCESS_WITH_INFO && er != SQL_NO_DATA {
		var sqlstate [6]byte
		var msg [SQL_MAX_MESSAGE_LENGTH]byte
		var errno SQLINTEGER
		var msglen SQLSMALLINT
		var i SQLSMALLINT
		SQLGetDiagRec := getFunc(dll, "SQLGetDiagRec")
		fmt.Fprintf(os.Stderr, "Error: %s\n", message)
		rc := SQLRETURN(utils.CfuncEbcdic(SQLGetDiagRec, uintptr(typ), uintptr(h), uintptr(1),
			uintptr(unsafe.Pointer(&sqlstate)),
			uintptr(unsafe.Pointer(&errno)),
			uintptr(unsafe.Pointer(&msg[0])),
			uintptr(len(msg)),
			uintptr(unsafe.Pointer(&msglen))))
		for i = 2; rc != SQL_NO_DATA; i++ {
			fmt.Fprintf(os.Stderr, "SQLSTATE: %s\n", string(sqlstate[:5]))
			fmt.Fprintf(os.Stderr, "error: %d\n", int(errno))
			fmt.Fprintf(os.Stderr, "%s\n", string(msg[:msglen]))
			rc = SQLRETURN(utils.CfuncEbcdic(SQLGetDiagRec, uintptr(typ), uintptr(h), uintptr(i),
				uintptr(unsafe.Pointer(&sqlstate)),
				uintptr(unsafe.Pointer(&errno)),
				uintptr(unsafe.Pointer(&msg[0])),
				uintptr(len(msg)),
				uintptr(unsafe.Pointer(&msglen))))
		}

	}
	return
}

func main() {
	foo()
}
