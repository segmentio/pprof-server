package pprofserver

import "html/template"

var listServices = htmlTemplate(`
<html>
<head>
	<title>Service List</title>
</head>
<body>
	<ul>{{ range . }}
		<li><a href="{{ .Href }}">{{ .Name }}</a></li>
	{{ end }}</ul>
</body>
</html>
`)

var lookupService = htmlTemplate(`
<html>
<head>
	<title>{{ .Name }}</title>
</head>
<body>
	<p>
		<a href="/services">&lt;&lt; Services</a>
	</p>
	<table>
		<tbody>{{ range .Profiles }}
			<tr>
				<td>{{ .Name }}</td>
				<td><a href="/tree{{ .Params }}">Tree</a></td>
				<td><a href="/flame{{ .Params }}">Flamegraph</a></td>
			</tr>
		{{ end }}</tbody>
	</table>
</body>
</html>
`)

var listNodes = htmlTemplate(`
<html>
<head>
	<title>{{ .Name }}</title>
</head>
<body>
	<p>
		<a href="/services">&lt;&lt; Services</a>
	</p>
	<table>
		<th>
			<tr>
				<td>Nodes</td>
			</tr>
		</th>
		<tbody>{{ range .Nodes }}
			<tr>
				<td><a href="{{ .Href }}">{{ .Endpoint }}</a></td>
			</tr>
		{{ end }}</tbody>
	</table>
</body>
</html>
`)

func htmlTemplate(s string) *template.Template {
	return template.Must(template.New("").Parse(s))
}
