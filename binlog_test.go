package stat

import (
	"encoding/hex"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

var (
	blPath     = ".binlog"
	testStore  *Store
	testBinLog *Binlog
)

func init() {
	os.RemoveAll(".testbinlog")
	os.RemoveAll(blPath)

	var e error
	testStore, e = NewStore(".testbinlog")
	if e != nil {
		panic(e)
	}

	testBinLog, e = NewWriteBinLog(blPath)
	if e != nil {
		panic(e)
	}
}

func TestBinLog(t *testing.T) {
	Convey("Test hitting", t, func() {
		h := NewHit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
		h.SetStore(testStore)
		bytes, e := h.AsBytes()
		So(e, ShouldEqual, nil)
		So(len(bytes), ShouldEqual, 36)
		Printf("%v\n", hex.EncodeToString(bytes))

		e = testBinLog.Append(bytes)
		So(e, ShouldEqual, nil)

		e = testBinLog.Append(bytes)
		So(e, ShouldEqual, nil)

		//e = testBinLog.Release()
		//So(e, ShouldEqual, nil)

		e = testBinLog.Archive("archive")
		So(e, ShouldEqual, nil)

		e = testBinLog.Append(bytes)
		So(e, ShouldEqual, nil)
	})
}

func BenchmarkAppendBinLog40(b *testing.B) {

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		h := NewHit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
		h.SetStore(testStore)
		bytes, _ := h.AsBytes()

		e := testBinLog.Append(bytes)
		if e != nil {
			fmt.Println(e)
		}
	}
}
