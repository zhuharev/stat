package stat

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestService(t *testing.T) {

	os.RemoveAll("testService")
	var testService *Service

	Convey("Test service", t, func() {

		Convey("test create service", func() {
			var e error
			testService, e = New("testService")
			So(e, ShouldBeNil)
			So(testService, ShouldNotBeNil)
			So(testService.Store, ShouldNotBeNil)
		})

		Convey("test hit unknown site", func() {
			e := testService.Hit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
			So(e, ShouldEqual, ErrUnknownSite)
		})

		Convey("add site", func() {
			e := testService.AddSite("google.com")
			So(e, ShouldBeNil)
		})

		Convey("test hit", func() {
			e := testService.Hit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
			So(e, ShouldBeNil)
		})

		Convey("score", func() {
			st, e := testService.Stat("google.com")
			So(e, ShouldBeNil)
			So(st.SiteId, ShouldBeGreaterThan, 0)
			So(st.TodayHit, ShouldBeGreaterThanOrEqualTo, 1)

			for i := 0; i < 100; i++ {
				e := testService.Hit("google.com", "/", "http://google.com", "1.1.1.1", "google chrome", "123123123213", "12321312312312312312312")
				So(e, ShouldBeNil)
			}

			st, e = testService.Stat("google.com")
			So(e, ShouldBeNil)
			So(st.TodayHit, ShouldBeGreaterThanOrEqualTo, 100)
		})
	})
}
