package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"github.com/fatih/color"
	"github.com/zhuharev/stat"
	"gopkg.in/macaron.v1"
)

var (
	statSrv *stat.Service
	// StatHost used testing host
	StatHost = "test.ru"
)

func init() {
	var e error
	statSrv, e = stat.New("data")
	if e != nil {
		panic(e)
	}
	go statSrv.ArchiveBinlogIfNeededEvery(time.Hour * 24 * 30)
}

func main() {
	m := newMacaron()
	m.Run(9000)
}

func handleHit(ctx *macaron.Context) {
	e := statSrv.HandleHit(ctx.Resp, ctx.Req.Request)
	if e != nil {
		color.Red("%s", e)
	}
	png1px := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAABGdBTUEAAK/INwWK6QAAAAtJREFUGFdj+A8EAAn7A/0r1QhFAAAAAElFTkSuQmCC"
	data, err := base64.StdEncoding.DecodeString(png1px)
	if err != nil {
		return
	}
	ctx.Resp.Header().Add("Content-Type", "image/png")
	fmt.Fprintf(ctx.Resp, "%s", data)
}

func handleStat(ctx *macaron.Context) {
	siteString := ctx.Query("site")
	if siteString == "" {
		u, e := url.Parse(ctx.Req.Referer())
		if e != nil {
			return
		}
		siteString = u.Host
	}
	stat, e := statSrv.Stat(siteString)
	if e != nil {
		color.Red("%s", e)
		return
	}

	svg := ctx.Query("format")
	if svg == "svg" || ctx.Req.URL.Path == "/stat.svg" {
		ctx.Resp.Header().Set("Content-type", "image/svg+xml")
		draw(stat.Rank, ctx.Resp)
		return
	}

	var resp = struct {
		Site      string
		HitCount  int64
		UniqCount int64
		Rank      int64
	}{
		ctx.Query("site"),
		stat.TodayHit,
		stat.TodayUniq,
		stat.Rank,
	}

	ctx.JSON(200, resp)
}
