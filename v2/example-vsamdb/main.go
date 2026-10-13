// Copyright 2026 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build zos
// +build zos

// example-vsamdb exercises the EzNoSQL (VSAMDB) interface of the go-recordio
// v2 module.  It creates a JSON database and a secondary index, writes three
// JSON documents, reads them back directly and sequentially, updates one of
// them and finally deletes it again.
//
// Before running the program, the system must provide:
//
//   - SMSVSAM running on the LPAR(s) that will share the database,
//   - an SMS storage class with a VSAM RLS CACHESET (-storclas),
//   - read/write access to the database data set name (-dsn).
//
// See example-vsamdb/README.md for the full list of prerequisites.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ibmruntimes/go-recordio/v2/eznosql"
)

// Book is the JSON document stored in the database.  Documents are UTF-8 JSON
// as they are stored in the VSAMDB data set, so Go structs are converted with
// encoding/json without any code page conversion.
type Book struct {
	ID         string   `json:"_id"`
	Title      string   `json:"Title"`
	Author     string   `json:"Author"`
	Year       int      `json:"Year,omitempty"`
	Characters []string `json:"Characters,omitempty"`
}

var (
	dsn       = flag.String("dsn", "HLQ.NOSQLDB", "database data set name (EBCDIC)")
	storclas  = flag.String("storclas", "", "SMS storage class with a VSAM RLS CACHESET (required)")
	primary   = flag.String("primary", "_id", "primary JSON key name")
	alternate = flag.String("alternate", "Author", "alternate JSON key name")
	maxSpace  = flag.Uint("maxspace", 5, "maximum database size in megabytes")
	keep      = flag.Bool("keep", false, "keep the database at the end instead of destroying it")
)

// quote returns the JSON quoted form of a key name or key value, the way the
// VSAMDB APIs expect them.
func quote(s string) string {
	return `"` + s + `"`
}

// fatal stops the program with the given error, if any.
func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// must stops the program when an EzNoSQL API call did not succeed.
func must(err *eznosql.Error) {
	if err == nil {
		return
	}
	log.Fatalf("%s (reason 0x%04x: %s)", err.API, uint16(err.Rsn), eznosql.ReasonText(int(err.Rsn)))
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: example-vsamdb -storclas STORCLAS [-dsn NAME] [-primary KEY]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *storclas == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Load the EzNoSQL C API DLL and bind the znsq_* entry points.
	if err := eznosql.Load(""); err != nil {
		log.Fatalf("load %s: %v", eznosql.DefaultDll, err)
	}
	defer eznosql.Unload()

	// The base name of the database.  Alternate index and path names are
	// derived from it, as the IBM sample program does.
	aixName := *dsn + ".AIX"
	pathName := *dsn + ".PATH"

	books := []Book{
		{ID: "01", Title: "The Fellowship of the Ring", Author: "J. R. R. Tolkien", Year: 1954},
		{ID: "02", Title: "The Two Towers", Author: "J. R. R. Tolkien", Year: 1954},
		{ID: "03", Title: "The Return of the King", Author: "J. R. R. Tolkien", Year: 1955},
	}

	// A stale database from a previous run would make the create fail with a
	// duplicate data set name, so remove it first and ignore the result.
	if err := eznosql.Destroy(*dsn); err != nil {
		log.Printf("note: destroy %s: %s (ignored)", *dsn, eznosql.ReasonText(int(err.Rsn)))
	}

	// ---- create the database and a secondary index -----------------------
	log.Printf("creating database %s", *dsn)
	must(eznosql.Create(*dsn, eznosql.CreateOptions{
		MaxSpace:   uint32(*maxSpace),
		PrimaryKey: quote(*primary),
		Storclas:   *storclas,
	}))

	idxFlags := uint32(eznosql.CreateIndexNonUnique)
	must(eznosql.CreateIndex(quote(*alternate), idxFlags, eznosql.CreateIndexOptions{
		BaseName: *dsn,
		AixName:  aixName,
		PathName: pathName,
		MaxSpace: uint32(*maxSpace),
	}))
	log.Printf("created index %s on key %q", aixName, *alternate)

	// ---- open a connection ----------------------------------------------
	con, err := eznosql.Open(*dsn, 0, eznosql.OpenOptions{
		Integrity: eznosql.IntegrityNRI,
		Timeout:   5,
	})
	must(err)
	defer eznosql.Close(con)
	log.Printf("connection to %s established", *dsn)

	// An index created above is inactive until it is added (built).
	must(eznosql.AddIndex(eznosql.AddIndexOptions{BaseName: *dsn, AixName: aixName}))
	log.Printf("index %s added", aixName)

	// ---- insert documents ------------------------------------------------
	for _, b := range books {
		doc := marshal(b)
		if _, err := eznosql.Write(con, string(doc), eznosql.WriteOptions{
			KeyName: quote(*primary),
		}); err != nil && err.Rsn != int32(eznosql.RsnMoreDuplicateKeys) {
			// RsnMoreDuplicateKeys is informational: the document was
			// inserted, only duplicate alternate key values follow.
			log.Fatalf("write %s: %s", b.ID, eznosql.ReasonText(int(err.Rsn)))
		}
		log.Printf("wrote document %s: %s", b.ID, string(doc))
	}

	// ---- direct read by primary key --------------------------------------
	doc, _, err := eznosql.Read(con, quote(*primary), quote("01"), eznosql.ReadOptions{})
	must(err)
	var book Book
	unmarshal(doc, &book)
	log.Printf("read by %s=%s: %+v", *primary, "01", book)

	// ---- sequential read over the secondary index ------------------------
	rs, err := eznosql.Position(con, quote(*alternate), quote(books[0].Author), eznosql.SearchEqual, 0)
	must(err)
	log.Printf("positioned on %s=%q", *alternate, books[0].Author)
	for {
		doc, err := eznosql.NextResult(con, &rs, eznosql.NextResultOptions{})
		// RsnMoreDuplicateKeys is informational: all three documents share
		// the same Author, so the code accompanies each document but the
		// last.
		if err != nil && err.Rsn != int32(eznosql.RsnMoreDuplicateKeys) {
			if err.Rsn == int32(eznosql.RsnEOD) {
				log.Printf("end of data reached")
				break
			}
			log.Fatalf("next_result: %s", eznosql.ReasonText(int(err.Rsn)))
		}
		var b Book
		unmarshal(doc, &b)
		log.Printf("next document %s: %q (%d)", b.ID, b.Title, b.Year)
	}
	must(eznosql.CloseResult(con, rs))

	// ---- update a document by primary key --------------------------------
	books[0].Title = "The Fellowship of the Ring (revised)"
	doc = marshal(books[0])
	must(eznosql.Update(con, string(doc), quote(*primary), quote("01")))
	log.Printf("document 01 updated to %q", books[0].Title)

	// ---- read for update through the secondary index ---------------------
	rs, err = eznosql.Position(con, quote(*alternate), quote(books[0].Author), eznosql.SearchEqual, 0)
	must(err)
	doc, err = eznosql.NextResult(con, &rs, eznosql.NextResultOptions{ForUpdate: true})
	must(err)
	updated := books[0]
	updated.Year = 1955
	doc = marshal(updated)
	must(eznosql.UpdateResult(con, rs, string(doc)))
	log.Printf("document 01 updated through result set, year %d", updated.Year)

	// ---- verify the update and delete the document -----------------------
	doc, _, err = eznosql.Read(con, quote(*primary), quote("01"), eznosql.ReadOptions{})
	must(err)
	var check Book
	unmarshal(doc, &check)
	log.Printf("document 01 is now %+v", check)

	must(eznosql.Delete(con, quote(*primary), quote("01")))
	log.Printf("document 01 deleted")

	if _, _, err := eznosql.Read(con, quote(*primary), quote("01"), eznosql.ReadOptions{}); err != nil {
		log.Printf("read of deleted document 01 reports 0x%04x (%s)",
			uint16(err.Rsn), eznosql.ReasonText(int(err.Rsn)))
	}

	// ---- statistics ------------------------------------------------------
	if report, err := eznosql.ReportStats(con); err == nil {
		log.Printf("report:\n%s", report)
	} else if err.Rsn != int32(eznosql.RsnUnknown) {
		log.Printf("report_stats failed: %s", eznosql.ReasonText(int(err.Rsn)))
	}

	// ---- close the connection -------------------------------------------
	must(eznosql.CloseResult(con, rs))
	must(eznosql.Close(con))
	log.Printf("connection closed")

	if *keep {
		log.Printf("database %s kept", *dsn)
		return
	}
	must(eznosql.Destroy(*dsn))
	log.Printf("database %s destroyed", *dsn)
}

// marshal converts v into the JSON document stored in the database, stopping
// the program when the conversion fails.
func marshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	fatal(err)
	return b
}

// unmarshal decodes a JSON document read from the database into v, stopping the
// program when the conversion fails.
func unmarshal(doc []byte, v interface{}) {
	fatal(json.Unmarshal(doc, v))
}
