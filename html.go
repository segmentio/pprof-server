package pprofserver

import "html/template"

var listServices = htmlTemplate("listServices", `
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

var lookupService = htmlTemplate("lookupService", `
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
				{{- if supportsFlamegraph .Params}}
				<td><a href="/tree{{ .Params }}">Tree</a></td>
				<td><a href="/flame{{ .Params }}">Flamegraph</a></td>
				{{- else}}
				<td><a href="/tree{{ .Params }}">Profile</a></td>
				{{- end}}
			</tr>
		{{ end }}</tbody>
	</table>
</body>
</html>
`)

var listNodes = htmlTemplate("listNodes", `
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

func htmlTemplate(name, s string) *template.Template {
	funcMap := template.FuncMap{
		"supportsFlamegraph": supportsFlamegraph,
	}
	tmpl := template.New(name).Funcs(funcMap)

	return template.Must(tmpl.Parse(s))
}
