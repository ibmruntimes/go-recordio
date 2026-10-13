# EzNoSQL (JSON VSAMDB) Example

An example of using the `eznosql` package to work with a JSON VSAMDB: a VSAM
KSDS data set defined with DATABASE(JSON) (z/OS 2.4+) that is accessed through
VSAM Record Level Sharing (RLS).  The program

1. creates a database and a secondary (alternate) index over the `Author` key,
2. opens a connection, builds the index,
3. writes three JSON documents (`encoding/json` structs, no code page
   conversion is needed: documents, key names and key values are UTF-8, the
   way they are stored in the data set),
4. reads one document back by its primary key `_id`,
5. scans the database sequentially through the secondary index until end of
   data,
6. updates a document directly by key and through a result set, and
7. deletes the document and reports the connection statistics.

## Prerequisites

* z/OS 2.4 or later with SMSVSAM running on every LPAR that will share the
  database.
* An SMS storage class with a VSAM RLS CACHESET, passed as `-storclas` (the
  only required flag).
* Read/write access for the user to the database data set name (`-dsn`), its
  alternate index cluster (`-dsn.AIX`) and its path (`-dsn.PATH`).
* The EzNoSQL C API DLL `libigwznsqd64.so` (shipped with z/OS) must be
  available in the program's link/STEPLIB concatenation; `eznosql.Load`
  resolves the `znsq_*` entry points from it.

## Running the example

```sh
go build
./example-vsamdb -storclas YOURCACHESET [-dsn HLQ.NOSQLDB] [-primary _id]
```

Add `-keep` to inspect the database with the IBM `igwvznsq` sample programs
instead of destroying it when the program ends.
