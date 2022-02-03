package pprofserver

import "html/template"

var listServices = htmlTemplate("listServices", `
<html>
<head>
	<title>Service List</title>
	<style>
		.service-list a {
			text-decoration: none;
		}
		
		/* 
		 * Alert box styles copied from Bootstrap, which has a MIT license:
		 * https://github.com/twbs/bootstrap/blob/main/LICENSE
		 */
		.alert {
			position: relative;
			padding: 0.75rem 1.25rem;
			margin-bottom: 1rem;
			border: 1px solid transparent;
			border-radius: 0.25rem;
		}

		.alert-danger {
			color: #721c24;
			background-color: #f8d7da;
			border-color: #f5c6cb;
		}
	</style>
</head>
<body>
	<div class="alert alert-danger">
		EKS services are not listed here.<br /><br />
		Use <a href="https://github.com/segmentio/kubectl-curl#collecting-profiles-of-go-programs">kubectl-curl</a> 
		to download profiles for EKS services.
	</div>
	<ul class="service-list">{{ range . }}
		<li><a title="{{ .Name }}" href="{{ .Href }}">{{ .Name }}</a></li>
	{{ end }}</ul>
</body>
</html>
`)

var lookupService = htmlTemplate("lookupService", `
<html>
<head>
	<title>{{ .Name }}</title>
	<style>a {text-decoration: none;}</style>
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
				<td><a href="/tree{{ .Params }}">🌲</a></td>
				<td><a href="/flame{{ .Params }}">🔥</a></td>
				{{- else}}
				<td><a href="/tree{{ .Params }}">📜</a></td>
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
	<style>a {text-decoration: none;}</style>
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
