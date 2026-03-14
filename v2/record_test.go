// Copyright 2021 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package zosrecordio

import (
	"testing"

	"github.com/ibmruntimes/go-recordio/v2/utils"
)

func TestFreadSIGILL(t *testing.T) {
	var line [32768]byte
	for i := 0; i < 5000; i++ {
		t.Logf("#%d iteration running...", i)
		fh := Fopen("//'SYS1.MACLIB(EXCP)'", "rb, lrecl=80, blksize=80, recfm=fb, type=record")
		if fh.Nil() {
			utils.Perror()
			t.Fatalf("#%d fopen failed", i)
			continue
		}
		bytes := fh.Fread(line[:])
		for bytes > 0 {
			bytes = fh.Fread(line[:])
		}
		if fh.Ferror() != nil {
			utils.Perror()
			t.Fatalf("#%d fread failed", i)
		}
		if !fh.Feof() {
			t.Fatalf("#%d read didn't receive EOF", i)
		}
		fh.Fclose()
	}
}
