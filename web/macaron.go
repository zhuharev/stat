package main

import (
	"fmt"
	"gopkg.in/macaron.v1"
)

func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())

	m.Get("/hit", handleHit)
	m.Get("/stat", handleStat)
	m.Get("/stat.svg", handleStat)
	m.Get("/html", handleHtml)
	m.Get("/", handleIndex)
	m.Post("/", handleIndexPost)

	return m
}

func handleHtml(ctx *macaron.Context) string {
	ctx.Resp.Header().Set("Content-type", "text/html")
	return fmt.Sprintf(`<a href="http://%[1]s/html">Site</a><img src='http://%[1]s/stat.svg'><script>
(new Image).src="http://%[1]s/hit?r"+escape(document.referrer)+";u"+escape(document.URL)+";"+Math.random()
</script>`, StatHost)
}

func handleIndex(ctx *macaron.Context) string {
	ctx.Resp.Header().Set("Content-type", "text/html")
	return `<form method="post">
	<input type="text" name="site" placeholder="host i.e example.com" />
		</form>`
}

func handleIndexPost(ctx *macaron.Context) string {
	ctx.Resp.Header().Set("Content-type", "text/html")
	site := ctx.Query("site")
	e := statSrv.AddSite(site)
	msg := fmt.Sprintf("Site %s added", site)
	if e != nil {
		msg = e.Error()
	}

	longMessage := `<br><textarea name="" id="" cols="30" rows="10">
<script>
(new Image).src="//%s/hit?r"+escape(document.referrer)+";u"+escape(document.URL)+";"+Math.random()
</script>
</textarea><br>
<a href="//%[1]s/stat?site=%[2]s">Json stat</a><br>
<a href="//%[1]s/stat?site=%[2]s&format=svg">Svg icon stat (rank)</a><br>
`
	return msg + fmt.Sprintf(longMessage, StatHost, site)
}
