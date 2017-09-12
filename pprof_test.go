package pprofserver

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePprofHome(t *testing.T) {
	r := strings.NewReader(`<html>
<head>
<title>/debug/pprof/</title>
</head>
<body>
/debug/pprof/<br>
<br>
profiles:<br>
<table>

<tr><td align=right>0<td><a href="block?debug=1">block</a>

<tr><td align=right>33<td><a href="goroutine?debug=1">goroutine</a>

<tr><td align=right>3<td><a href="heap?debug=1">heap</a>

<tr><td align=right>0<td><a href="mutex?debug=1">mutex</a>

<tr><td align=right>11<td><a href="threadcreate?debug=1">threadcreate</a>

</table>
<br>
<a href="goroutine?debug=2">full goroutine stack dump</a><br>
</body>
</html>
`)

	p, err := parsePprofHome(r)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(p, []profile{
		{Name: "block", URL: "block?debug=1"},
		{Name: "goroutine", URL: "goroutine?debug=1"},
		{Name: "heap", URL: "heap?debug=1"},
		{Name: "mutex", URL: "mutex?debug=1"},
		{Name: "threadcreate", URL: "threadcreate?debug=1"},
		{Name: "full goroutine stack dump", URL: "goroutine?debug=2"},
	}) {
		t.Error(p)
	}
}
