package pprofserver

import (
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

/*
This function parses the output of the /debug/pprof endpoint of services, which
looks like this:

<html>
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
*/
func parsePprofHome(r io.Reader) ([]profile, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var profiles []profile
	var search func(*html.Node)
	search = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			var p profile

			for _, a := range n.Attr {
				if a.Key == "href" {
					p.URL = a.Val
					break
				}
			}

			if n.FirstChild != nil {
				p.Name = n.FirstChild.Data
			}

			profiles = append(profiles, p)
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			search(c)
		}
	}
	search(doc)
	return profiles, nil
}
