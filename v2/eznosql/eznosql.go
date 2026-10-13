// Copyright 2026 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build zos
// +build zos

// Package eznosql gives Go programs access to VSAMDB (EzNoSQL) databases: VSAM
// KSDS data sets defined with DATABASE(JSON)/DATABASE(BSON) (z/OS 2.4+) that are
// accessed through VSAM Record Level Sharing (RLS).
//
// The package binds the znsq_* C API shipped with z/OS in libigwznsqd64.so
// (AMODE 64) and libigwznsqd31.so (AMODE 31).  Because a Go program on z/OS runs
// in AMODE 64, the 64-bit DLL is used by default.  No cgo is involved: the entry
// points are resolved with dlopen()/dlsym() through
// github.com/ibmruntimes/go-recordio/v2/utils and are called with
// runtime.CallLeFuncByPtr(), like the record I/O wrappers of
// github.com/ibmruntimes/go-recordio/v2.
//
// Encoding rules (z/OS DFSMS "Using VSAM record-level sharing"):
//
//   - Data set names, alternate index names, path names and SMS class names are
//     passed as EBCDIC C strings, the same way the C samples pass them.
//   - JSON documents, JSON key names and key values are UTF-8 byte strings, the
//     way they are stored in the VSAMDB data set.  A key name and a key value
//     are JSON quoted, e.g. key name "\"_id\"" and key value "\"01\"", so that
//     documents can be converted to and from Go values with encoding/json
//     without any code page conversion.
//   - The C API supports the JSON format only (igwvznsq.h: enum znsq_format
//     { JSON_DB = 0 }).  BSON data sets are processed through the VSAM record
//     management macros (ACB/RPL), not through this API.
package eznosql

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

// DefaultDll is the EzNoSQL C API DLL used when Load is given an empty name.  It
// matches the AMODE 64 addressing mode of a Go program on z/OS.
const DefaultDll = "libigwznsqd64.so"

// Every EzNoSQL API entry point bound by Load.
var apiNames = []string{
	"znsq_create",
	"znsq_create_index",
	"znsq_destroy",
	"znsq_add_index",
	"znsq_drop_index",
	"znsq_report_stats",
	"znsq_open",
	"znsq_close",
	"znsq_read",
	"znsq_position",
	"znsq_next_result",
	"znsq_close_result",
	"znsq_write",
	"znsq_write_result",
	"znsq_delete",
	"znsq_delete_result",
	"znsq_update",
	"znsq_update_result",
	"znsq_commit",
	"znsq_abort",
	"znsq_set_autocommit",
	"znsq_last_result",
}

// Reason codes returned in the low order 16 bits of an API return value
// (znsq_err(e) == e & 0xFFFF in igwvznsq.h), as documented in the IBM EzNoSQL
// documentation.  Reason codes X'34', X'35' and X'36' are informational: the
// document was still returned or written, and the caller may continue.
const (
	RsnSuccess              = 0x0000 // accompanies a return code of 0
	RsnDsnDup               = 0x0002 // duplicate database name (znsq_create)
	RsnDsnInv               = 0x0003 // invalid database name (znsq_create)
	RsnNoSpace              = 0x0004 // no candidate space or volumes available
	RsnNoGuaranteedSpace    = 0x0005 // no guaranteed storage space
	RsnNoStorageSpace       = 0x0006 // no storage space for max_space
	RsnParamRange           = 0x0007 // IDCAMS parameter out of range
	RsnNoLogstream          = 0x0008 // LOG_ALL without a logstream ID
	RsnDsnNotFound          = 0x0009 // database not found (znsq_destroy)
	RsnErrBufTooSmall       = 0x000A // internal error buffer too small
	RsnDupIndexKey          = 0x000C // duplicate key name for a unique index
	RsnIndexActive          = 0x000D // add_index for an already active index
	RsnDsnInUse             = 0x000E // database in use, SYSDSN not obtainable
	RsnDsnNotFoundOpen      = 0x0012 // database not found (open/close)
	RsnDsnComponentNotFound = 0x0013 // database component not found
	RsnPathNotFound         = 0x0014 // path entry not found (read/write)
	RsnAltKeyNotFound       = 0x0015 // alternate key not found (open/close/read)
	RsnUndoLogFailed        = 0x001B // undo log unavailable or write failed
	RsnFwdLogFailed         = 0x001C // forward recovery log write failed
	RsnCfCacheFailed        = 0x001D // CF cache structure failure
	RsnCfCacheUnavailable   = 0x001E // CF cache structure unavailable
	RsnCfCacheSetUnavail    = 0x001F // CF cache set not found in storage class
	RsnPrimaryKeyNotInCat   = 0x0033 // primary key not found in catalog
	RsnMoreDuplicateKeys    = 0x0034 // non-unique index: more duplicates follow
	RsnTruncatedKey         = 0x0035 // key truncated to 251 bytes
	RsnNuxTruncatedKey      = 0x0036 // both of the above
	RsnDsnNotFoundAccess    = 0x0037 // database not found (first access)
	RsnNotAuthorized        = 0x0038 // user not authorized to access database
	RsnEmptyInputOpen       = 0x0039 // open for input on an empty database
	RsnDbDamaged            = 0x0041 // high used RBA beyond high allocated
	RsnWriteForceFailed     = 0x0042 // forced update of a duplicate failed
	RsnEOD                  = 0x0043 // end of data (sequential read)
	RsnDuplicateKey         = 0x0044 // duplicate primary/alternate key inserted
	RsnKeyNotFound          = 0x0045 // key name/key value pair not found
	RsnIncompatibleLock     = 0x0046 // same UR requested the key incompatibly
	RsnDeadlock             = 0x0047 // deadlock time out
	RsnLockTimeout          = 0x0048 // document lock time out
	RsnOutOfSpace           = 0x0049 // write could not extend the database
	RsnServerUnavailable    = 0x0079 // SMSVSAM not available or a new instance (121)
	RsnMemoryError          = 0x0050 // internal memory obtain failed
	RsnBufferTooSmall       = 0x0051 // buffer too small, required size returned
	RsnRplReuse             = 0x0052 // internal RPL reuse error
	RsnMaxRequests          = 0x0053 // over 1024 concurrent requests
	RsnUpdInvKey            = 0x0054 // update would change the primary key
	RsnNoKeyName            = 0x0055 // document lacks the primary key name
	RsnIncompleteDoc        = 0x0056 // document has an incomplete length
	RsnRplError             = 0x0057 // invalid internal RPL options
	RsnCrossMemory          = 0x0058 // API issued in cross-memory mode
	RsnIndexCorrupt         = 0x0059 // database index is corrupted
	RsnReadOnlyAccess       = 0x005A // write issued on a read-only database
	RsnRetainedLock         = 0x005B // retained lock from a failed transaction
	RsnIncompleteSegment    = 0x0060 // document has an incomplete segment
	RsnNoBaseRecord         = 0x0061 // alternate index out of sync with base
	RsnMaxDuplicateKeys     = 0x0062 // maximum duplicate alternate keys reached
	RsnIndexWriteCorrupt    = 0x0063 // corrupt index entry about to be written
	RsnDbPotentiallyDmgd    = 0x0064 // database potentially damaged
	RsnNotAuthorizedAlter   = 0x0065 // update authority required (add/drop index)
	RsnMultipleAddIndex     = 0x0066 // another add_index is in progress
	RsnIoError              = 0x0067 // physical I/O error while writing
	RsnBackwardInsert       = 0x0068 // sequential write in backward mode
	RsnNewServerInstance    = 0x0069 // new server instance after open
	RsnNoPosition           = 0x006A // next/close_result without position
	RsnRetainedLockHeld     = 0x006B // retained lock at close time
	RsnKeySequence          = 0x006C // sequential write with a lower key
	RsnQuiesced             = 0x0070 // database quiesced for backup
	RsnTvsDisabling         = 0x0071 // transactional server is disabling
	RsnInvalidTask          = 0x0072 // not the task that opened the database
	RsnDocTooLargeForLog    = 0x0073 // document over 62K in a recoverable db
	RsnNoPriorReadForUpdate = 0x0074 // update/delete without a read for update
	RsnIndexNotActive       = 0x0075 // access of an inactive index
	RsnAbend                = 0x0077 // system ABEND during the API
	RsnBtreeNotFound        = 0x0078 // BTREE control block not found
	RsnVrgbNotFound         = 0x007A // SMSVSAM did not initialize (VRGB)
	RsnTvsUnavailable       = 0x007C // DFSMSTVS not available (recoverable db)
	RsnTvsNotInstalled      = 0x007D // DFSMSTVS not installed (recoverable db)
	RsnConBlockFailed       = 0x0080 // connection control block could not be built
	RsnRsBlockFailed        = 0x0081 // result set control block delete failed
	RsnEmptyDB              = 0x008A // position attempted for an empty database
	RsnInternalLogic        = 0x0090 // internal system logic error
	RsnFlushIoError         = 0x0092 // I/O error flushing buffers at close
	RsnSecIndexNotFound     = 0x0096 // secondary index not found (add_index)
	RsnPrimIndexNotFound    = 0x0097 // primary index not found (add_index)
	RsnMemoryShortage       = 0x0098 // storage for an internal block failed
	RsnParamError           = -1     // parameter error (return code -1)
	RsnUnknown              = 0x7FFF // reason code not available
)

// znsq_format.  Only JSON is supported by the C API.
const FormatJSON = 0

// znsq_log_options: database recovery (RLS logging) options.
const (
	LogNone = 0 // non-recoverable
	LogUndo = 2 // backout logging
	LogAll  = 3 // backout and forward recovery logging (needs a log stream)
)

// znsq_integrity: read integrity, which decides whether shared locks are taken
// and how long they are held.
const (
	IntegrityNRI = 0 // no read integrity (default)
	IntegrityCR  = 1 // consistent read
	IntegrityCRE = 2 // consistent read extended
)

// znsq_boolean: three state boolean.
const (
	BoolDefault = 0 // not set, use the database default
	BoolTrue    = 1
	BoolFalse   = 2
)

// znsq_search_method: how znsq_position interprets the key value.
const (
	SearchEqual       = 0 // position at the first (or last) key value equal to
	SearchGreaterThan = 1 // position at the first (or last) key value >=
)

// Flag bits for znsq_create_index.
const (
	CreateIndexNonUnique     = 1 << 0 // non-unique alternate index
	CreateIndexDescendingKey = 1 << 1 // descending sequential access
)

// Flag bits for znsq_open.
const (
	OpenReadOnly       = 1 << 0 // read only connection
	OpenForceUpdate    = 1 << 1 // replace duplicates instead of failing
	OpenDescendingKeys = 1 << 2 // descending sequential access
)

// Connection is the znsq_connection_t token returned by Open.
type Connection int64

// ResultSet is the znsq_result_set_t token returned by Position.
type ResultSet int32

// Nil reports whether the connection token is valid.
func (c Connection) Nil() bool { return c == 0 }

// Nil reports whether the result set token is valid.
func (r ResultSet) Nil() bool { return r == 0 }

// Error describes a failed EzNoSQL API call.  Ret holds the full return code,
// Rsn the API reason code (Ret & 0xFFFF).
type Error struct {
	API string // API name, e.g. "znsq_read"
	Ret int32  // full return code as returned by the API
	Rsn int32  // reason code, see the Rsn* constants
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s failed: return code %d, reason 0x%04x (%s)",
		e.API, e.Ret, uint16(e.Rsn), ReasonText(int(e.Rsn)))
}

// ReasonText maps an EzNoSQL reason code to a short readable description.
func ReasonText(rsn int) string {
	switch rsn {
	case RsnSuccess:
		return "success"
	case RsnDsnDup:
		return "duplicate database name (znsq_create without a destroy)"
	case RsnDsnInv:
		return "invalid database name"
	case RsnNoSpace:
		return "no candidate space or volumes available"
	case RsnNoGuaranteedSpace:
		return "no guaranteed storage space for max_space"
	case RsnNoStorageSpace:
		return "no storage space available for max_space"
	case RsnParamRange:
		return "IDCAMS parameter out of range"
	case RsnNoLogstream:
		return "LOG_ALL without a forward recovery log stream"
	case RsnDsnNotFound:
		return "database not found (znsq_destroy)"
	case RsnErrBufTooSmall:
		return "internal error buffer too small"
	case RsnDupIndexKey:
		return "duplicate key name for a unique index (znsq_add_index)"
	case RsnIndexActive:
		return "index already active (znsq_add_index, use znsq_drop_index first)"
	case RsnDsnInUse:
		return "database in use, cannot obtain SYSDSN (znsq_create_index)"
	case RsnDsnNotFoundOpen:
		return "database not found (open/close)"
	case RsnDsnComponentNotFound:
		return "database component not found"
	case RsnPathNotFound:
		return "path entry not found"
	case RsnAltKeyNotFound:
		return "alternate key not found"
	case RsnUndoLogFailed:
		return "undo log unavailable or write failed"
	case RsnFwdLogFailed:
		return "forward recovery log write failed"
	case RsnCfCacheFailed:
		return "CF cache structure failure"
	case RsnCfCacheUnavailable:
		return "CF cache structure unavailable"
	case RsnCfCacheSetUnavail:
		return "CF cache set not found in the storage class"
	case RsnPrimaryKeyNotInCat:
		return "JSON primary key not found in the catalog"
	case RsnMoreDuplicateKeys:
		return "more duplicate key values follow"
	case RsnTruncatedKey:
		return "alternate key value truncated to 251 bytes"
	case RsnNuxTruncatedKey:
		return "truncated key and more duplicate key values follow"
	case RsnDsnNotFoundAccess:
		return "database not found (first access)"
	case RsnNotAuthorized:
		return "user not authorized to access the database"
	case RsnEmptyInputOpen:
		return "open for input attempted on an empty database"
	case RsnDbDamaged:
		return "database damaged: high used RBA above high allocated"
	case RsnWriteForceFailed:
		return "forced update of a duplicate document failed"
	case RsnEOD:
		return "end of data"
	case RsnDuplicateKey:
		return "duplicate primary or alternate key value"
	case RsnKeyNotFound:
		return "key name/key value not found"
	case RsnIncompatibleLock:
		return "incompatible lock requested by the same unit of recovery"
	case RsnDeadlock:
		return "deadlock time out"
	case RsnLockTimeout:
		return "document lock time out"
	case RsnOutOfSpace:
		return "out of space, the database could not be extended"
	case RsnServerUnavailable:
		return "SMSVSAM is not available or is a new instance"
	case RsnMemoryError:
		return "internal memory obtain failed"
	case RsnBufferTooSmall:
		return "buffer too small, required size returned"
	case RsnRplReuse:
		return "internal RPL reuse error"
	case RsnMaxRequests:
		return "maximum number of concurrent requests exceeded"
	case RsnUpdInvKey:
		return "update would change the primary key value"
	case RsnNoKeyName:
		return "document does not contain the primary key name"
	case RsnIncompleteDoc:
		return "incomplete document found"
	case RsnRplError:
		return "invalid RPL options supplied by the system"
	case RsnCrossMemory:
		return "API issued in cross-memory mode"
	case RsnIndexCorrupt:
		return "database index is corrupted"
	case RsnReadOnlyAccess:
		return "write issued against a read-only database"
	case RsnRetainedLock:
		return "retained lock from a failed transaction"
	case RsnIncompleteSegment:
		return "incomplete document segment found"
	case RsnNoBaseRecord:
		return "alternate index out of sync with the base database"
	case RsnMaxDuplicateKeys:
		return "maximum number of duplicate alternate keys reached"
	case RsnIndexWriteCorrupt:
		return "corrupt index entry about to be written"
	case RsnDbPotentiallyDmgd:
		return "database potentially damaged, writes are failed"
	case RsnNotAuthorizedAlter:
		return "user not authorized to alter the database (add/drop index)"
	case RsnMultipleAddIndex:
		return "another znsq_add_index is in progress"
	case RsnIoError:
		return "physical I/O error while writing"
	case RsnBackwardInsert:
		return "sequential write in backward mode"
	case RsnNewServerInstance:
		return "new server instance after open"
	case RsnNoPosition:
		return "next_result/close_result without a preceding position"
	case RsnRetainedLockHeld:
		return "retained lock held at close time"
	case RsnKeySequence:
		return "sequential write with a key below the previous key"
	case RsnQuiesced:
		return "database quiesced for backup"
	case RsnTvsDisabling:
		return "transactional server instance is disabling"
	case RsnInvalidTask:
		return "request from a task that did not open the database"
	case RsnDocTooLargeForLog:
		return "document larger than 62K in a recoverable database"
	case RsnNoPriorReadForUpdate:
		return "update or delete without a prior read for update"
	case RsnIndexNotActive:
		return "index not active (inactivated by znsq_drop_index)"
	case RsnAbend:
		return "system ABEND during the API"
	case RsnBtreeNotFound:
		return "BTREE control block not found, invalid connection"
	case RsnVrgbNotFound:
		return "VRGB control block not found, SMSVSAM is not active"
	case RsnTvsUnavailable:
		return "DFSMSTVS not available (recoverable database)"
	case RsnTvsNotInstalled:
		return "DFSMSTVS not installed (recoverable database)"
	case RsnConBlockFailed:
		return "connection control block could not be built"
	case RsnRsBlockFailed:
		return "result set control block could not be deleted"
	case RsnEmptyDB:
		return "position attempted for an empty database"
	case RsnInternalLogic:
		return "internal system logic error"
	case RsnFlushIoError:
		return "I/O error flushing buffers at close"
	case RsnSecIndexNotFound:
		return "secondary index not found (znsq_add_index)"
	case RsnPrimIndexNotFound:
		return "primary index not found (znsq_add_index)"
	case RsnMemoryShortage:
		return "storage for an internal control block failed"
	case RsnUnknown:
		return "unknown reason code"
	case RsnParamError:
		return "parameter error"
	}
	return fmt.Sprintf("reason 0x%04x", uint16(rsn))
}

var dll utils.Dll
var api map[string]uintptr

// Load opens the EzNoSQL DLL and binds every API used by this package.  An
// empty name selects DefaultDll.
func Load(name string) error {
	if name == "" {
		name = DefaultDll
	}
	d := utils.Dll{}
	if err := d.Open(name); err != nil {
		return fmt.Errorf("eznosql: %s: %w", name, err)
	}
	funcs := make(map[string]uintptr, len(apiNames))
	for _, f := range apiNames {
		fp, err := d.Sym(f)
		if err != nil {
			d.Close()
			return fmt.Errorf("eznosql: %s: %w", name, err)
		}
		funcs[f] = fp
	}
	dll = d
	api = funcs
	return nil
}

// Unload releases the EzNoSQL DLL.
func Unload() error {
	if !dll.Valid {
		return nil
	}
	err := dll.Close()
	dll = utils.Dll{}
	api = nil
	if err != nil {
		return fmt.Errorf("eznosql: %w", err)
	}
	return nil
}

// notLoaded builds the parameter error returned when Load has not been called.
func notLoaded(name string) *Error {
	return &Error{API: name, Ret: -1, Rsn: RsnParamError}
}

// sigset is a z/OS signal set (see github.com/ibmruntimes/go-recordio/v2).
type sigset uint64

// sigblock_SIGURG blocks SIGURG only.  SIGURG is the Go runtime asynchronous
// preemption signal on z/OS and an LE call must not be interrupted by it.
var sigblock_SIGURG = sigset(1<<63) >> (syscall.SIGURG - 1)

const (
	sigBlock   = 0
	sigSetmask = 2
)

// disableAsyncPreempt locks the OS thread and blocks SIGURG while an LE call is
// in flight.  It mirrors the helper used by the record I/O wrappers in
// github.com/ibmruntimes/go-recordio/v2, where an asynchronous preemption signal
// arriving inside an LE call abended the program.
func disableAsyncPreempt() func() {
	runtime.LockOSThread()
	var origset sigset
	runtime.EnterSyscall()
	utils.Clib(utils.SYS_PTHREAD_SIGMASK, uintptr(sigBlock),
		uintptr(unsafe.Pointer(&sigblock_SIGURG)), uintptr(unsafe.Pointer(&origset)))
	runtime.ExitSyscall()
	return func() {
		runtime.EnterSyscall()
		utils.Clib(utils.SYS_PTHREAD_SIGMASK, uintptr(sigSetmask),
			uintptr(unsafe.Pointer(&origset)), uintptr(unsafe.Pointer(nil)))
		runtime.ExitSyscall()
		runtime.UnlockOSThread()
	}
}

// call invokes an EzNoSQL API function.  The thread is switched to EBCDIC mode
// for the duration of the call and asynchronous preemption is blocked, the
// combination this repository uses for LE calls that take EBCDIC parameters
// (data set names, SMS classes).
func call(name string, args ...uintptr) int32 {
	cleanup := disableAsyncPreempt()
	defer cleanup()
	utils.ThreadEbcdicMode()
	defer utils.ThreadAsciiMode()
	return int32(uint32(utils.Cfunc(api[name], args...)))
}

// check turns an API return value into an *Error.
func check(apiName string, ret int32) *Error {
	if ret == 0 {
		return nil
	}
	return &Error{API: apiName, Ret: ret, Rsn: ret & 0xffff}
}

// informational reports whether a reason code is informational: the document
// was returned or written and only the condition is reported (X'34': more
// duplicate alternate key values follow, X'35'/X'36': the alternate key was
// truncated to 251 bytes).  The data APIs return the document together with
// the *Error in that case, so the caller can use it and continue.
func informational(ret int32) bool {
	switch ret & 0xffff {
	case RsnMoreDuplicateKeys, RsnTruncatedKey, RsnNuxTruncatedKey:
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Option structures.  The layouts below match the C structures in
// /usr/include/zos/igwvznsq.h compiled for AMODE 64 (LP64, one byte enums).  The
// compile time assertions make a build fail if this package ever drifts from the
// C header.
// ---------------------------------------------------------------------------

// CreateOptions describes a database to be created by Create.  Storclas is the
// only required member; it must name an SMS storage class that carries a VSAM
// RLS CACHESET.
type CreateOptions struct {
	MaxSpace      uint32 // maximum database space in megabytes (required)
	AvgDocSize    uint32 // average document size in bytes (0 = system default)
	UpdatePercent uint32 // percentage of updates relative to writes
	LogOptions    int    // LogNone (default), LogUndo or LogAll
	PrimaryKey    string // UTF-8 JSON key name, e.g. "\"_id\"" ("" = implicit key)
	Storclas      string // SMS storage class with a VSAM RLS CACHESET
	Mgmtclas      string // SMS management class (optional)
	Dataclas      string // SMS data class (optional)
	LogstreamID   string // forward recovery log stream, required for LogAll
	OrderedIndex  bool   // ordered primary index (needs OA64954)
}

// cCreateOptions is the C znsq_create_options structure.
type cCreateOptions struct {
	Version       uint8 // 1, or 2 when the ordered index option is used
	Format        uint8 // enum znsq_format, JSON only
	_             uint16
	MaxSpace      uint32
	AvgDocSize    uint32
	UpdatePercent uint32
	LogOptions    uint8 // enum znsq_log_options
	_             [7]byte
	PrimaryKey    uintptr // UTF-8 JSON key name, or NULL
	Storclas      uintptr // EBCDIC
	Mgmtclas      uintptr // EBCDIC, or NULL
	Dataclas      uintptr // EBCDIC, or NULL
	LogstreamID   uintptr // EBCDIC, or NULL
	OrderedIndex  uint32
}

const _ = uint(unsafe.Sizeof(cCreateOptions{})) - 72

// CreateIndexOptions describes a secondary (alternate) index to be created by
// CreateIndex.
type CreateIndexOptions struct {
	BaseName      string // primary database name
	AixName       string // name of the alternate index cluster
	PathName      string // name of the path used to open the alternate index
	Dataclas      string // SMS data class (optional)
	MaxSpace      uint32 // maximum index space in megabytes
	UpdatePercent uint32 // percentage of updates relative to writes
	Integrity     int    // IntegrityNRI (default), IntegrityCR or IntegrityCRE
}

// cCreateIndexOptions is the C znsq_create_index_options structure.
type cCreateIndexOptions struct {
	Version       uint8
	_             [7]byte
	BaseName      uintptr // EBCDIC
	AixName       uintptr // EBCDIC
	PathName      uintptr // EBCDIC
	Dataclas      uintptr // EBCDIC, or NULL
	MaxSpace      uint32
	UpdatePercent uint32
	Integrity     uint8 // enum znsq_integrity
	_             [3]byte
}

const _ = uint(unsafe.Sizeof(cCreateIndexOptions{})) - 56

// AddIndexOptions builds and activates an index created by CreateIndex.
type AddIndexOptions struct {
	BaseName string // primary database name
	AixName  string // alternate index cluster name
}

// cAddIndexOptions is the C znsq_add_index_options structure.
type cAddIndexOptions struct {
	Version  uint8
	_        [7]byte
	BaseName uintptr // EBCDIC
	AixName  uintptr // EBCDIC
}

const _ = uint(unsafe.Sizeof(cAddIndexOptions{})) - 24

// DropIndexOptions disables an alternate index.
type DropIndexOptions struct {
	AixName string // alternate index cluster name
}

// cDropIndexOptions is the C znsq_drop_index_options structure.
type cDropIndexOptions struct {
	Version uint8
	_       [7]byte
	AixName uintptr // EBCDIC
}

const _ = uint(unsafe.Sizeof(cDropIndexOptions{})) - 16

// OpenOptions describes the connection established by Open.
type OpenOptions struct {
	Integrity  int   // IntegrityNRI (default), IntegrityCR or IntegrityCRE
	Timeout    int16 // lock wait timeout in seconds (0 = do not wait)
	Autocommit int   // BoolDefault (default), BoolTrue or BoolFalse
}

// cOpenOptions is the C znsq_open_options structure.
type cOpenOptions struct {
	Version    uint8
	Integrity  uint8 // enum znsq_integrity
	Timeout    int16
	Autocommit uint8 // enum znsq_boolean
	_          [1]byte
}

const _ = uint(unsafe.Sizeof(cOpenOptions{})) - 6

// ReadOptions is the C znsq_read_options structure.  ForUpdate takes a document
// level exclusive lock and returns the result set token that UpdateResult,
// DeleteResult and CloseResult expect.
type ReadOptions struct {
	ForUpdate bool
}

// NextResultOptions carries the znsq_next_result flags.  ForUpdate returns a
// token that stays valid for UpdateResult or DeleteResult.
type NextResultOptions struct {
	ForUpdate bool
}

// cReadOptions is the C znsq_read_options structure.
type cReadOptions struct {
	ResultSet int32
}

const _ = uint(unsafe.Sizeof(cReadOptions{})) - 4

// WriteOptions is the C znsq_write_options structure.  KeyName is the UTF-8 JSON
// key name of the primary index.  AutokeyBytes is the size of the buffer that
// receives the internally generated key of a keyless database.
type WriteOptions struct {
	KeyName      string
	AutokeyBytes int
}

// cWriteOptions is the C znsq_write_options structure.
type cWriteOptions struct {
	Version       uint8
	_             [7]byte
	KeyName       uintptr // UTF-8 JSON key name, or NULL
	AutokeyBuffer uintptr // buffer for a generated (implicit) key, or NULL
	AutokeyLength uint32
}

const _ = uint(unsafe.Sizeof(cWriteOptions{})) - 32

// WriteResultOptions is the C znsq_write_result_options structure.
type WriteResultOptions struct {
	KeyName      string
	AutokeyBytes int
}

// cWriteResultOptions is the C znsq_write_result_options structure.
type cWriteResultOptions struct {
	Version       uint8
	_             [7]byte
	KeyName       uintptr // UTF-8 JSON key name, or NULL
	AutokeyBuffer uintptr // buffer for a generated (implicit) key, or NULL
	AutokeyLength uint32
}

const _ = uint(unsafe.Sizeof(cWriteResultOptions{})) - 32

// CommitOptions is the C znsq_commit_options structure used by SetAutocommit.
type CommitOptions struct {
	Autocommit int // 0 = off, 1 = on
}

// cCommitOptions is the C znsq_commit_options structure.
type cCommitOptions struct {
	Version    uint8
	AutoCommit uint8 // enum znsq_autocommit
}

const _ = uint(unsafe.Sizeof(cCommitOptions{})) - 2

// ---------------------------------------------------------------------------
// C string helpers.  Data set and SMS class names are EBCDIC; JSON key names,
// key values and documents are UTF-8 and are passed unchanged.
// ---------------------------------------------------------------------------

// ebcdic returns s as a null terminated EBCDIC string in a buffer of its own.
func ebcdic(s string) []byte {
	if s == "" {
		return nil
	}
	b := make([]byte, len(s))
	copy(b, s)
	utils.AtoE(b)
	return append(b, 0)
}

// utf8 returns s as a null terminated UTF-8 string in a buffer of its own.
func utf8(s string) []byte {
	if s == "" {
		return nil
	}
	return append([]byte(s), 0)
}

// ptr returns the address of b, or NULL for an empty buffer.
func ptr(b []byte) uintptr {
	if len(b) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&b[0]))
}

// ---------------------------------------------------------------------------
// Database management
// ---------------------------------------------------------------------------

// Create creates an EzNoSQL database with the given data set name.
func Create(dsname string, opts CreateOptions) *Error {
	if api == nil {
		return notLoaded("znsq_create")
	}
	dsn := ebcdic(dsname)
	storclas := ebcdic(opts.Storclas)
	mgmtclas := ebcdic(opts.Mgmtclas)
	dataclas := ebcdic(opts.Dataclas)
	logstream := ebcdic(opts.LogstreamID)
	key := utf8(opts.PrimaryKey)

	c := cCreateOptions{
		Format:        FormatJSON,
		MaxSpace:      opts.MaxSpace,
		AvgDocSize:    opts.AvgDocSize,
		UpdatePercent: opts.UpdatePercent,
		PrimaryKey:    ptr(key),
		Storclas:      ptr(storclas),
		Mgmtclas:      ptr(mgmtclas),
		Dataclas:      ptr(dataclas),
		LogstreamID:   ptr(logstream),
	}
	c.Version = 1
	if opts.OrderedIndex {
		c.Version = 2 // ordered_index is only usable from version 2 on
		c.OrderedIndex = 1
	}
	switch opts.LogOptions {
	case LogUndo, LogAll:
		c.LogOptions = uint8(opts.LogOptions)
	}
	ret := call("znsq_create", ptr(dsn), uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(dsn)
	runtime.KeepAlive(storclas)
	runtime.KeepAlive(mgmtclas)
	runtime.KeepAlive(dataclas)
	runtime.KeepAlive(logstream)
	runtime.KeepAlive(key)
	runtime.KeepAlive(&c)
	return check("znsq_create", ret)
}

// CreateIndex creates a secondary index over an alternate key name.  The index
// is created inactive; call AddIndex to build and activate it.  A secondary
// index can only be created while the database is closed.
func CreateIndex(altkey string, flags uint32, opts CreateIndexOptions) *Error {
	if api == nil {
		return notLoaded("znsq_create_index")
	}
	akey := utf8(altkey)
	base := ebcdic(opts.BaseName)
	aix := ebcdic(opts.AixName)
	path := ebcdic(opts.PathName)
	dataclas := ebcdic(opts.Dataclas)

	c := cCreateIndexOptions{
		BaseName:      ptr(base),
		AixName:       ptr(aix),
		PathName:      ptr(path),
		Dataclas:      ptr(dataclas),
		MaxSpace:      opts.MaxSpace,
		UpdatePercent: opts.UpdatePercent,
	}
	switch opts.Integrity {
	case IntegrityCR, IntegrityCRE:
		c.Integrity = uint8(opts.Integrity)
	}
	ret := call("znsq_create_index", ptr(akey), uintptr(flags), uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(akey)
	runtime.KeepAlive(base)
	runtime.KeepAlive(aix)
	runtime.KeepAlive(path)
	runtime.KeepAlive(dataclas)
	runtime.KeepAlive(&c)
	return check("znsq_create_index", ret)
}

// Destroy deletes a database, including its alternate indexes.
func Destroy(dsname string) *Error {
	if api == nil {
		return notLoaded("znsq_destroy")
	}
	dsn := ebcdic(dsname)
	ret := call("znsq_destroy", ptr(dsn))
	runtime.KeepAlive(dsn)
	return check("znsq_destroy", ret)
}

// AddIndex builds and activates an alternate index created by CreateIndex.
func AddIndex(opts AddIndexOptions) *Error {
	if api == nil {
		return notLoaded("znsq_add_index")
	}
	base := ebcdic(opts.BaseName)
	aix := ebcdic(opts.AixName)
	c := cAddIndexOptions{BaseName: ptr(base), AixName: ptr(aix)}
	ret := call("znsq_add_index", uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(base)
	runtime.KeepAlive(aix)
	runtime.KeepAlive(&c)
	return check("znsq_add_index", ret)
}

// DropIndex disables an alternate index.  The index can be re-activated, and
// rebuilt, with AddIndex.
func DropIndex(opts DropIndexOptions) *Error {
	if api == nil {
		return notLoaded("znsq_drop_index")
	}
	aix := ebcdic(opts.AixName)
	c := cDropIndexOptions{AixName: ptr(aix)}
	ret := call("znsq_drop_index", uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(aix)
	runtime.KeepAlive(&c)
	return check("znsq_drop_index", ret)
}

// LastResult returns the diagnostic text of the last failed API call, as
// returned by znsq_last_result() in EBCDIC, converted to ASCII.
func LastResult() (string, *Error) {
	if api == nil {
		return "", notLoaded("znsq_last_result")
	}
	size := uint64(4096)
	buf := make([]byte, size)
	ret := call("znsq_last_result", uintptr(unsafe.Pointer(&size)), ptr(buf))
	if ret != int32(RsnSuccess) {
		need := size
		if need == 0 || need > 1<<28 {
			return "", check("znsq_last_result", ret)
		}
		size = need
		buf = make([]byte, size+1)
		ret = call("znsq_last_result", uintptr(unsafe.Pointer(&size)), ptr(buf))
		runtime.KeepAlive(buf)
		runtime.KeepAlive(&size)
		if e := check("znsq_last_result", ret); e != nil {
			return "", e
		}
	}
	runtime.KeepAlive(buf)
	runtime.KeepAlive(&size)
	utils.EtoA(buf[:size])
	return string(buf[:size]), nil
}

// ---------------------------------------------------------------------------
// Connections
// ---------------------------------------------------------------------------

// Open establishes a connection to a database.  The connection must be closed
// by the same task that opened it, so prefer "con, err := Open(...)" together
// with "defer eznosql.Close(con)".
func Open(dsname string, flags uint32, opts OpenOptions) (Connection, *Error) {
	if api == nil {
		return 0, notLoaded("znsq_open")
	}
	dsn := ebcdic(dsname)
	var con int64
	c := cOpenOptions{Timeout: opts.Timeout}
	switch opts.Integrity {
	case IntegrityCR, IntegrityCRE:
		c.Integrity = uint8(opts.Integrity)
	}
	switch opts.Autocommit {
	case BoolTrue, BoolFalse:
		c.Autocommit = uint8(opts.Autocommit)
	}
	ret := call("znsq_open", uintptr(unsafe.Pointer(&con)), ptr(dsn), uintptr(flags),
		uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(dsn)
	runtime.KeepAlive(&con)
	runtime.KeepAlive(&c)
	if e := check("znsq_open", ret); e != nil {
		return 0, e
	}
	return Connection(con), nil
}

// Close closes a connection opened by Open.
func Close(con Connection) *Error {
	if api == nil {
		return notLoaded("znsq_close")
	}
	return check("znsq_close", call("znsq_close", uintptr(con)))
}

// ReportStats returns the JSON statistics report of a connection.
func ReportStats(con Connection) (string, *Error) {
	if api == nil {
		return "", notLoaded("znsq_report_stats")
	}
	size := uint64(8192)
	buf := make([]byte, size)
	ret := call("znsq_report_stats", uintptr(con), uintptr(unsafe.Pointer(&size)), ptr(buf))
	if ret != int32(RsnSuccess) {
		// Grow the buffer to the size the API reports and re-drive once.
		need := size
		if need == 0 || need > 1<<28 {
			return "", check("znsq_report_stats", ret)
		}
		size = need
		buf = make([]byte, size+1)
		ret = call("znsq_report_stats", uintptr(con), uintptr(unsafe.Pointer(&size)), ptr(buf))
		runtime.KeepAlive(buf)
		runtime.KeepAlive(&size)
		if e := check("znsq_report_stats", ret); e != nil {
			return "", e
		}
	}
	runtime.KeepAlive(buf)
	runtime.KeepAlive(&size)
	return string(buf[:size]), nil
}

// Commit ends the current transaction and releases the document level locks.
func Commit(con Connection) *Error {
	if api == nil {
		return notLoaded("znsq_commit")
	}
	return check("znsq_commit", call("znsq_commit", uintptr(con)))
}

// Abort backs out the current transaction, restores the original documents and
// releases the document level locks.
func Abort(con Connection) *Error {
	if api == nil {
		return notLoaded("znsq_abort")
	}
	return check("znsq_abort", call("znsq_abort", uintptr(con)))
}

// SetAutocommit turns auto commit on or off.  Turning it on commits a
// transaction that is still open.
func SetAutocommit(con Connection, opts CommitOptions) *Error {
	if api == nil {
		return notLoaded("znsq_set_autocommit")
	}
	c := cCommitOptions{}
	if opts.Autocommit == 1 {
		c.AutoCommit = 1
	}
	ret := call("znsq_set_autocommit", uintptr(con), uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(&c)
	return check("znsq_set_autocommit", ret)
}

// ---------------------------------------------------------------------------
// Document retrieval
// ---------------------------------------------------------------------------

// Read retrieves the document identified by key and keyValue.  key is the UTF-8
// JSON key name (with the quotes) of a primary index or of an active secondary
// index, keyValue the matching JSON value (with the quotes).  The document is
// returned as the raw UTF-8 JSON, ready for encoding/json.
//
// Read grows its buffer when the API reports it to be too small, so callers do
// not have to size it.  When opts.ForUpdate is set, the document is locked and
// the returned result set token can be passed to UpdateResult, DeleteResult or
// CloseResult.  Read can return a document together with an informational
// *Error (RsnMoreDuplicateKeys, RsnTruncatedKey): the document is valid and
// only the condition is reported.
func Read(con Connection, key, keyValue string, opts ReadOptions) ([]byte, ResultSet, *Error) {
	if api == nil {
		return nil, 0, notLoaded("znsq_read")
	}
	k := utf8(key)
	v := utf8(keyValue)
	flags := uint32(0)
	if opts.ForUpdate {
		flags = 1 // VSAMDB_READ_UPDATE
	}
	bufLen := uint64(4096)
	for attempt := 0; ; attempt++ {
		buf := make([]byte, bufLen)
		var ro cReadOptions
		ret := call("znsq_read", uintptr(con), ptr(buf), uintptr(unsafe.Pointer(&bufLen)),
			ptr(k), ptr(v), uintptr(flags), uintptr(unsafe.Pointer(&ro)))
		runtime.KeepAlive(buf)
		runtime.KeepAlive(k)
		runtime.KeepAlive(v)
		runtime.KeepAlive(&ro)
		if ret == int32(RsnSuccess) {
			return buf[:bufLen], ResultSet(ro.ResultSet), nil
		}
		if informational(ret) {
			return buf[:bufLen], ResultSet(ro.ResultSet), check("znsq_read", ret)
		}
		if ret&0xffff == int32(RsnBufferTooSmall) && attempt == 0 && usable(&bufLen) {
			continue
		}
		return nil, 0, check("znsq_read", ret)
	}
}

// Position locates a key value in a primary or in an active secondary index and
// returns a result set token for the sequential reads that follow.  An empty
// keyValue positions at the first (or last) document of the index.  Finish the
// scan with CloseResult.
func Position(con Connection, key, keyValue string, method int, flags uint32) (ResultSet, *Error) {
	if api == nil {
		return 0, notLoaded("znsq_position")
	}
	k := utf8(key)
	v := utf8(keyValue)
	var rs int32
	ret := call("znsq_position", uintptr(con), uintptr(unsafe.Pointer(&rs)), ptr(k),
		ptr(v), uintptr(method), uintptr(flags))
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
	runtime.KeepAlive(&rs)
	if e := check("znsq_position", ret); e != nil {
		return 0, e
	}
	return ResultSet(rs), nil
}

// NextResult reads the next document of the scan opened by Position and
// advances the result set token.  ReadForUpdate returns a token that stays valid
// for UpdateResult or DeleteResult.  Like Read, NextResult can return a
// document together with an informational *Error (RsnMoreDuplicateKeys,
// RsnTruncatedKey); RsnEOD ends the scan.
func NextResult(con Connection, rs *ResultSet, opts NextResultOptions) ([]byte, *Error) {
	if api == nil {
		return nil, notLoaded("znsq_next_result")
	}
	if rs == nil {
		return nil, &Error{API: "znsq_next_result", Ret: -1, Rsn: RsnParamError}
	}
	flags := uint32(0)
	if opts.ForUpdate {
		flags = 1 // VSAMDB_NEXT_RESULT_FLAG_UPDATE
	}
	bufLen := uint64(4096)
	for attempt := 0; ; attempt++ {
		buf := make([]byte, bufLen)
		ret := call("znsq_next_result", uintptr(con), uintptr(unsafe.Pointer(rs)), ptr(buf),
			uintptr(unsafe.Pointer(&bufLen)), uintptr(flags))
		runtime.KeepAlive(buf)
		runtime.KeepAlive(rs)
		runtime.KeepAlive(&bufLen)
		if ret == int32(RsnSuccess) {
			return buf[:bufLen], nil
		}
		if informational(ret) {
			return buf[:bufLen], check("znsq_next_result", ret)
		}
		if ret&0xffff == int32(RsnBufferTooSmall) && attempt == 0 && usable(&bufLen) {
			continue
		}
		return nil, check("znsq_next_result", ret)
	}
}

// CloseResult ends the scan started by Position and releases the result set.
func CloseResult(con Connection, rs ResultSet) *Error {
	if api == nil {
		return notLoaded("znsq_close_result")
	}
	return check("znsq_close_result", call("znsq_close_result", uintptr(con), uintptr(rs)))
}

// usable reports whether the length the API just wrote, after reporting that the
// buffer was too small, can be used as the size of the next buffer.
func usable(n *uint64) bool {
	v := *n
	if v == 0 || v > 1<<28 {
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Document management
// ---------------------------------------------------------------------------

// Write inserts a new JSON document.  opts.KeyName names the primary key of the
// database (UTF-8, with the quotes).  Set opts.AutokeyBytes to have an
// internally generated (implicit) key returned in the first result.  A write
// into a non-unique secondary index can return an informational *Error
// (RsnMoreDuplicateKeys) together with the auto-generated key: the document
// was inserted.
func Write(con Connection, doc string, opts WriteOptions) ([]byte, *Error) {
	if api == nil {
		return nil, notLoaded("znsq_write")
	}
	buf := utf8(doc)
	key := utf8(opts.KeyName)
	c := cWriteOptions{
		KeyName:       ptr(key),
		AutokeyLength: uint32(opts.AutokeyBytes),
	}
	var autokey []byte
	if opts.AutokeyBytes > 0 {
		autokey = make([]byte, opts.AutokeyBytes)
		c.AutokeyBuffer = ptr(autokey)
	}
	ret := call("znsq_write", uintptr(con), ptr(buf), uintptr(len(doc)),
		uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(buf)
	runtime.KeepAlive(key)
	runtime.KeepAlive(autokey)
	runtime.KeepAlive(&c)
	if e := check("znsq_write", ret); e != nil {
		if informational(ret) {
			return autokey, e
		}
		return nil, e
	}
	return autokey, nil
}

// WriteResult inserts a new document while positioned in a scan.
func WriteResult(con Connection, rs *ResultSet, doc string, opts WriteResultOptions) ([]byte, *Error) {
	if api == nil {
		return nil, notLoaded("znsq_write_result")
	}
	buf := utf8(doc)
	key := utf8(opts.KeyName)
	c := cWriteResultOptions{
		KeyName:       ptr(key),
		AutokeyLength: uint32(opts.AutokeyBytes),
	}
	var autokey []byte
	if opts.AutokeyBytes > 0 {
		autokey = make([]byte, opts.AutokeyBytes)
		c.AutokeyBuffer = ptr(autokey)
	}
	ret := call("znsq_write_result", uintptr(con), uintptr(unsafe.Pointer(rs)), ptr(buf),
		uintptr(len(doc)), uintptr(unsafe.Pointer(&c)))
	runtime.KeepAlive(buf)
	runtime.KeepAlive(key)
	runtime.KeepAlive(autokey)
	runtime.KeepAlive(&c)
	runtime.KeepAlive(rs)
	if e := check("znsq_write_result", ret); e != nil {
		if informational(ret) {
			return autokey, e
		}
		return nil, e
	}
	return autokey, nil
}

// Update replaces the document identified by key and keyValue.  The primary key
// value of a document cannot change (RsnInvalidUpdate).
func Update(con Connection, doc, key, keyValue string) *Error {
	if api == nil {
		return notLoaded("znsq_update")
	}
	buf := utf8(doc)
	k := utf8(key)
	v := utf8(keyValue)
	ret := call("znsq_update", uintptr(con), ptr(buf), ptr(k), ptr(v))
	runtime.KeepAlive(buf)
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
	return check("znsq_update", ret)
}

// UpdateResult replaces a document that Read or NextResult locked for update.
func UpdateResult(con Connection, rs ResultSet, doc string) *Error {
	if api == nil {
		return notLoaded("znsq_update_result")
	}
	buf := utf8(doc)
	ret := call("znsq_update_result", uintptr(con), uintptr(rs), ptr(buf), uintptr(len(doc)))
	runtime.KeepAlive(buf)
	return check("znsq_update_result", ret)
}

// Delete erases the document identified by key and keyValue.
func Delete(con Connection, key, keyValue string) *Error {
	if api == nil {
		return notLoaded("znsq_delete")
	}
	k := utf8(key)
	v := utf8(keyValue)
	ret := call("znsq_delete", uintptr(con), ptr(k), ptr(v))
	runtime.KeepAlive(k)
	runtime.KeepAlive(v)
	return check("znsq_delete", ret)
}

// DeleteResult erases a document that Read or NextResult locked for update.
func DeleteResult(con Connection, rs ResultSet) *Error {
	if api == nil {
		return notLoaded("znsq_delete_result")
	}
	return check("znsq_delete_result", call("znsq_delete_result", uintptr(con), uintptr(rs)))
}
