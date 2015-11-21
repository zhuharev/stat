package stat

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	os.RemoveAll("testStore")
	Convey("test store", t, func() {

		Convey("test open stores", func() {
			_, e := NewStore("testStore")
			So(e, ShouldEqual, nil)
		})

	})
}
