# Scope
Record I/O in Go: a Go module for record I/O in VSAM databases directly from Go (no need for cgo)

# Updates 2024-04-16
New utility is available to create Go-style structures from HLASM ADATA file.
Install by:
`go install github.com/IBM/godsect@latest`

# Updates
New updates to this package is in the v2 directory, the current directory is left unchanged for compatibility.

# Usage

THe following interfaces are defined:

```
// Fopen takes a name and mode as byte slices and returns a RecordStream.
// Note that the strings encoded in thee slices must be null terminated.
// The utility function ConvertStringToSlice automatically null-terminates.
func Fopen(fname []byte, mode []byte) (rs RecordStream)


// Freopen behaves the same as Fopen, but takes a previously used RecordStream
func Freopen(fname []byte, mode []byte, rs RecordStream) (rso RecordStream)


// Flocate locates a record by Key, received as a byte slice
// Returns 0 if successful, otherwise EOF
func  (rs RecordStream) Flocate(key []byte, options int) int


// Fread reads a record.
// If the buffer is not big enough, the record will be truncated.
// The actual number of bytes read is returned
func  (rs RecordStream) Fread(buffer []byte) int


// Fdelrec deletes the last read record
// Returns 0 if successful, otherwise non-zero
func  (rs RecordStream) Fdelrec() int


// Feof returns true if the last set EOF flag value is EOF
func  (rs RecordStream) Feof() bool


// Ferror returns the last set error value
func  (rs RecordStream) Ferror() error


// Fupdate updates the last read record to be the new record value in buffer
// It returns the size of the updated record
func  (rs RecordStream) Fupdate(buffer []byte) int


//Fwrite writes one record contained in buffer to the rs stream.
// It returns the number of bytes written.
// Note, the size of the record is the size of the slice.
func  (rs RecordStream) Fwrite(buffer []byte) int


// Fclose closes the stream. Returns 0 if successful, otherwise EOF.
func  (rs RecordStream) Fclose() int

```

Utility Functions

When doing record i/o one would typically use a language structure to represent the record.
In C one would use a struct, and then cast the struct to and from a byte array as needed to
use as arguments to the various functions in the language environment. In Go, that kind of
casting isn't normally available, so some similar conversions routines are provided here.

```
// ConvertStringToSlice copies the string into the given slice.
// Always includeds a null terminator in the copy.
// Returns a new empty slice if the string doesn't fit.
func ConvertStringToSlice(s string, bi []byte) (bo []byte)


// ConvertStructToSlice returns a byte slice that shares storage
// with the incoming struct. The length of the slice will be
// exactly the size of the struct.
func ConvertStructToSlice(i interface{}) (slice []byte)


// ConvertSliceToStruc returns an interface that can be type asserted to be the
// same type as the incoming pointer to struct "i".
// The pointer that results from such a type assertion will share storage
// with the incoming byte slice "bi". The struct size is the second returned value.
// Thus the following sequence leaves buffp as a pointer to a FixedHeader struct
// that shares storage with myBigSlice[:buffSize]:
// var buffp *FixedHeader_T
// buffp_, buffSize := zosrecordio.ConvertSliceToStruct(buffp, myBigSlice)
// buffp = buffp_.(*FixedHeader_T)
// Note: if the  incoming slice isn't big enough, it returns <nil, 0>.
func ConvertSliceToStruct(i interface{}, bi []byte) (interface{}, int)
```


Example

In the example directory is a program that can be built with "go build test.go". It exercises nearly all the interfaces using a struct inspired by the blog "VSAM: The no-charge z/OS DB" by Mike Fulton (https://makingdeveloperslivesbetter.wordpress.com/2021/03/17/vsam-the-no-charge-z-os-db/). To run this example program, provide an argument with the name of the KEY.PATH dataset for your VSAM cluster. For example, if your cluster is HLQ.TESTDB then the program would be invoked as:

```
./test "//'HLQ.TESTDB.KEY.PATH'"
```

You can create such a VSAM database using the crtvsamxsysvar script from the source: 
```
git@github.com:MikeFultonDev/samples.git, branch Xsysvar.

```
Note: to use this you also need Z Open Automation Utilities (ZOAU) installed.
