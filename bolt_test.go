package stat

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBolt(t *testing.T) {
	Convey("Test it", t, func() {

		var (
			host  = "zhuahrev.ru"
			host2 = "golang.org"
			host3 = "google.com"
		)

		_, e := testStore.InsertAutoInc("sites", host)
		So(e, ShouldEqual, nil)

		id, e := testStore.Get("sites", host)
		So(e, ShouldEqual, nil)
		So(id, ShouldNotEqual, 0)

		ok, e := testStore.HasSite(host)
		So(e, ShouldEqual, nil)
		So(ok, ShouldEqual, true)

		_, e = testStore.InsertAutoInc("sites", host2)
		So(e, ShouldEqual, nil)

		id, e = testStore.Get("sites", host2)
		So(e, ShouldEqual, nil)
		So(id, ShouldBeGreaterThan, 2)

		id, e = testStore.GetOrInsert("sites", host)
		So(e, ShouldEqual, nil)
		So(id, ShouldBeGreaterThan, 0)

		id, e = testStore.GetOrInsert("sites", host3)
		So(e, ShouldEqual, nil)
		So(id, ShouldNotEqual, 0)
	})

}
