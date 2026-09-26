# go-recordio

Record I/O in Go: a Go module for record I/O in VSAM databases directly from Go (no need for cgo). z/OS only.

## Which version should I use? v2.

All active development happens in `v2/`. New users should use `v2` only.

* Current: `github.com/ibmruntimes/go-recordio/v2` — source in `v2/`, docs in `v2/README.md`, examples in `v2/example-*/`
* Legacy, frozen: `github.com/ibmruntimes/go-recordio` — source in the repo root (`record.go`, `utils/`, `example/`). Kept only for backward compatibility, no new updates.

### Why is the root frozen and everything new goes into `v2/`?

This is Go modules convention for major versions >= v2. The module path must include the major-version suffix (`/v2`), so `v2/` has its own `go.mod`:

* `go.mod`: `module github.com/ibmruntimes/go-recordio`
* `v2/go.mod`: `module github.com/ibmruntimes/go-recordio/v2`

Freezing the root lets existing v1 users keep building without breaking changes, while v2 can evolve independently side-by-side in the same repo.

### Usage

For new code:

```bash
go get github.com/ibmruntimes/go-recordio/v2
```

```go
import zosrecordio "github.com/ibmruntimes/go-recordio/v2"
```

See `v2/README.md` and `v2/example-recordio/` for current API docs and examples.

For existing v1 code only — no changes needed, but no new features either:

```go
import zosrecordio "github.com/ibmruntimes/go-recordio"
```

### Repo layout

* `v2/record.go`, `v2/utils/`, `v2/record_test.go` — current code
* `v2/example-*/` — current examples (recordio, cobol, db2, dvm, etc.)
* `record.go`, `utils/`, `example/` — legacy v1 code, frozen
* `v2/` has its own `LICENSE`, `CONTRIBUTING.md`, `MAINTAINERS.md`

## Tools

Utility to create Go-style structures from an HLASM ADATA file:

```bash
go install github.com/IBM/godsect@latest
```

## Appendix: legacy v1 API (frozen, for reference only)

The following interfaces are defined in the root `record.go`:

```
// Fopen takes a name and mode as strings and returns a RecordStream.
func Fopen(fname string, mode string) (rs RecordStream)

// Freopen behaves the same as Fopen, but takes a previously used RecordStream
func Freopen(fname string, mode string, rs RecordStream) (rso RecordStream)

// Flocate locates a record by Key, received as a byte slice
// Returns 0 if successful, otherwise EOF
func (rs RecordStream) Flocate(key []byte, options LocOptions) int

// Fread reads a record.
// If the buffer is not big enough, the record will be truncated.
// The actual number of bytes read is returned
func (rs RecordStream) Fread(buffer []byte) int

// Fdelrec deletes the last read record
// Returns 0 if successful, otherwise non-zero
func (rs RecordStream) Fdelrec() int

// Feof returns the last set EOF flag value
func (rs RecordStream) Feof() bool

// Ferror returns the last set error value
func (rs RecordStream) Ferror() error

// Fupdate updates the last read record to be the new record value in buffer
// It returns the size of the updated record
func (rs RecordStream) Fupdate(buffer []byte) int

// Fwrite writes one record contained in buffer to the rs stream.
// It returns the number of bytes written.
// Note, the size of the record is the size of the slice.
func (rs RecordStream) Fwrite(buffer []byte) int

// Fclose closes the stream. Returns 0 if successful, otherwise EOF.
func (rs RecordStream) Fclose() int
```

Utility Functions

When doing record I/O one would typically use a language structure to represent the record.
In C one would use a struct, and then cast the struct to and from a byte array as needed to
use as arguments to the various functions in the language environment. In Go, that kind of
casting isn't normally available, so some similar conversion routines are provided here.

```
// ConvertStringToSlice copies the string into the given slice.
// Always includes a null terminator in the copy.
// Returns a new empty slice if the string doesn't fit.
func ConvertStringToSlice(s string, bi []byte) (bo []byte)

// ConvertStructToSlice returns a byte slice that shares storage
// with the incoming struct. The length of the slice will be
// exactly the size of the struct.
func ConvertStructToSlice(i interface{}) (slice []byte)

// ConvertSliceToStruct returns an interface that can be type asserted to be the
// same type as the incoming pointer to struct "i".
// The pointer that results from such a type assertion will share storage
// with the incoming byte slice "bi". The struct size is the second returned value.
// Thus the following sequence leaves buffp as a pointer to a FixedHeader struct
// that shares storage with myBigSlice[:buffSize]:
// var buffp *FixedHeader_T
// buffp_, buffSize := zosrecordio.ConvertSliceToStruct(buffp, myBigSlice)
// buffp = buffp_.(*FixedHeader_T)
// Note: if the incoming slice isn't big enough, it returns <nil, 0>.
func ConvertSliceToStruct(i interface{}, bi []byte) (interface{}, int)
```

Example (legacy v1)

In the `example/` directory is a program that can be built with `go build test.go`. It exercises nearly all the interfaces using a struct inspired by the blog "VSAM: The no-charge z/OS DB" by Mike Fulton (https://makingdeveloperslivesbetter.wordpress.com/2021/03/17/vsam-the-no-charge-z-os-db/). To run this example program, provide an argument with the name of the KEY.PATH dataset for your VSAM cluster. For example, if your cluster is HLQ.TESTDB then the program would be invoked as:

```
./test "//'HLQ.TESTDB.KEY.PATH'"
```

You can create such a VSAM database using the crtvsamxsysvar script from the source:
```
git@github.com:MikeFultonDev/samples.git, branch Xsysvar.
```
Note: to use this you also need Z Open Automation Utilities (ZOAU) installed.

For current examples, see `v2/example-recordio/` and the other `v2/example-*/` directories.
