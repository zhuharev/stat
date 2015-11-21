package main

import (
	"fmt"
	"github.com/ajstarks/svgo"
	"io"
)

func draw(num int64, w io.Writer) {
	width := 88
	height := 31
	inString := "In"
	cityName := "Rating"

	canvas := svg.New(w)
	canvas.Start(width, height)
	canvas.LinearGradient("a", 0, 0, 0, 100, []svg.Offcolor{{Color: "#bbb", Offset: 0, Opacity: 0.1}, {Color: "#bbb", Offset: 100, Opacity: 0.1}})
	canvas.Roundrect(0, 0, width, height, 2, 2, "fill:#5272B4;")
	canvas.Roundrect(0, 0, 28, height, 2, 2, "fill:#555;")

	strcnt := fmt.Sprint(num)
	fsize := "18"
	if num > 9 {
		fsize = "15"
	}
	if num > 99 {
		fsize = "12"
	}
	if num > 999 {
		fsize = "10"
	}
	if num > 9999 {
		fsize = "8"
		strcnt = "9999+"
	}
	canvas.Text(14, height/2+7, strcnt, "text-anchor:middle;font-size:"+fsize+"px;fill:white;font-family:DejaVu Sans,Verdana,Geneva,sans-serif")
	canvas.Text(58, height/2+10, cityName, "width:60px;text-anchor:middle;font-size:10px;fill:white;font-family:DejaVu Sans,Verdana,Geneva,sans-serif")
	canvas.Text(59, height/2, inString, "width:60px;text-anchor:middle;font-size:10px;fill:white;font-family:DejaVu Sans,Verdana,Geneva,sans-serif")
	canvas.End()
}
