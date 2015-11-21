package stat

import (
	"encoding/hex"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestHit(t *testing.T) {
	Convey("Test hitting", t, func() {
		h := NewHit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
		h.SetStore(testStore)
		bytes, e := h.AsBytes()
		So(e, ShouldEqual, nil)
		So(len(bytes), ShouldEqual, 36)
		Printf("%v\n", hex.EncodeToString(bytes))
	})
}
