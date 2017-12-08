package pprofserver

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"sort"
	"strings"

	"github.com/segmentio/events"
	"github.com/segmentio/events/log"
	"github.com/segmentio/objconv/json"
)

type Handler struct {
	Prefix   string
	Registry Registry
	Client   *http.Client
}

type Registry interface {
	ListServices(ctx context.Context) ([]string, error)
	LookupService(ctx context.Context, name string) ([]string, error)
}

func (h *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	header := res.Header()
	header.Set("Content-Language", "en")
	header.Set("Server", "pprof-server")

	switch req.Method {
	case http.MethodGet, http.MethodHead:
	default:
		http.Error(res, "only GET and HEAD are allowed", http.StatusMethodNotAllowed)
		return
	}

	switch path := req.URL.Path; {
	case path == "/", path == "/services":
		h.serveRedirect(res, req, "/services/")

	case path == "/tree":
		h.serveTree(res, req)

	case path == "/flame":
		h.serveFlame(res, req)

	case path == "/services/":
		h.serveListServices(res, req)

	case strings.HasPrefix(path, "/services/"):
		h.serveLookupService(res, req)

	default:
		h.serveNotFound(res, req)
	}
}

func (h *Handler) serveRedirect(res http.ResponseWriter, req *http.Request, url string) {
	http.Redirect(res, req, url, http.StatusFound)
}

func (h *Handler) serveNotFound(res http.ResponseWriter, req *http.Request) {
	http.NotFound(res, req)
}

func (h *Handler) serveListServices(res http.ResponseWriter, req *http.Request) {
	var services []service

	if h.Registry != nil {
		names, err := h.Registry.ListServices(req.Context())
		if err != nil {
			events.Log("error listing services: %{error}s", err)
		}
		services = make([]service, 0, len(names))
		for _, name := range names {
			services = append(services, service{
				Name: name,
				Href: "/services/" + name,
			})
		}
	}

	render(res, req, listServices, services)
}

func (h *Handler) serveLookupService(res http.ResponseWriter, req *http.Request) {
	var ctx = req.Context()
	var name = strings.TrimPrefix(path.Clean(req.URL.Path), "/services/")
	var srv service

	if h.Registry != nil {
		endpoints, _ := h.Registry.LookupService(ctx, name)
		profchan := make(chan []profile, len(endpoints))

		for _, ep := range endpoints {
			go func(ep string) {
				p, err := h.fetchService(ctx, name, ep)
				profchan <- p
				if err != nil {
					events.Log("error fetching service profiles of %{service}s (%{endpoint}s): %{error}s", name, ep, err)
				}
			}(ep)
		}

		for range endpoints {
			p := <-profchan
			srv.Profiles = append(srv.Profiles, p...)
		}
	}

	if len(srv.Profiles) == 0 {
		p, err := h.fetchService(ctx, name, name) // try with DNS
		if err != nil {
			events.Log("error fetching service profiles of %{service}s: %{error}s", name, err)
		} else {
			srv.Profiles = append(srv.Profiles, p...)
		}
	}

	sort.Slice(srv.Profiles, func(i int, j int) bool {
		p1 := srv.Profiles[i]
		p2 := srv.Profiles[j]

		if p1.Name != p2.Name {
			return p1.Name < p2.Name
		}

		return p1.URL < p2.URL
	})

	srv.Name = name
	srv.Href = "/services/" + name
	render(res, req, lookupService, srv)
}

func (h *Handler) serveFlame(res http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	serviceURL := queryString.Get("url")
	queryString.Del("url")

	if len(serviceURL) == 0 {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	rawRes, err := h.client().Get(serviceURL)
	if err != nil {
		res.WriteHeader(http.StatusBadGateway)
		events.Log("error querying %{url}s: %{error}s", serviceURL, err)
		return
	}
	defer rawRes.Body.Close()

	if err := renderFlamegraph(res, rawRes.Body, serviceURL, query2pprofArgs(queryString)); err == nil {
		return
	}

	// failed to render a graph; fall back to serving the raw profile
	h.serveRawProfile(res, req, serviceURL)

}

func (h *Handler) serveTree(res http.ResponseWriter, req *http.Request) {
	queryString := req.URL.Query()
	serviceURL := queryString.Get("url")
	queryString.Del("url")

	if len(serviceURL) == 0 {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	args := []string{
		"tool",
		"pprof",
		"-svg",
		"-symbolize",
		"remote",
	}

	args = append(args, query2pprofArgs(queryString)...)
	args = append(args, serviceURL)

	buffer := &bytes.Buffer{}
	buffer.Grow(32768)

	pprof := exec.CommandContext(req.Context(), "go", args...)
	pprof.Stdin = nil
	pprof.Stdout = buffer
	pprof.Stderr = log.NewWriter("", 0, events.DefaultHandler)

	if pprof.Run() == nil {
		buffer.WriteTo(res)
		return
	}

	// failed to render a graph; fall back to serving the raw profile
	h.serveRawProfile(res, req, serviceURL)
}

func query2pprofArgs(q url.Values) (args []string) {
	for flag, values := range q {
		if len(values) == 0 {
			args = append(args, "-"+flag)
		} else {
			for _, value := range values {
				args = append(args, "-"+flag, value)
			}
		}
	}
	return
}

func (h *Handler) serveRawProfile(w http.ResponseWriter, r *http.Request, url string) {
	res, err := h.client().Get(url)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		events.Log("error querying %{url}s: %{error}s", url, err)
		return
	}
	io.Copy(w, res.Body)
	res.Body.Close()
}

func (h *Handler) fetchService(ctx context.Context, name string, endpoint string) (prof []profile, err error) {
	var req *http.Request
	var res *http.Response
	var baseURL = endpoint + h.prefix()

	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	if req, err = http.NewRequest(http.MethodGet, baseURL, nil); err != nil {
		return
	}

	if res, err = h.client().Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	if prof, err = parsePprofHome(res.Body); err != nil {
		return
	}

	// For some reason the default profiles aren't returned by the /debug/pprof/
	// home page.
	prof = append(prof,
		profile{
			Name: "profile",
			URL:  "profile?seconds=5",
		},
		profile{
			Name: "trace",
			URL:  "trace?seconds=5",
		},
	)

	for i, p := range prof {
		// For heap profiles, inject the options for capturing the allocated objects
		// or the allocated space.
		if path, _ := splitPathQuery(p.URL); path == "heap" {
			prof[i].Name = p.Name + " (objects in use)"
			prof[i].URL = baseURL + p.URL
			prof[i].Params = "?inuse_objects&url=" + url.QueryEscape(prof[i].URL)

			prof = append(prof,
				profile{
					Name:   p.Name + " (space in use)",
					URL:    baseURL + p.URL,
					Params: "?inuse_space&url=" + url.QueryEscape(prof[i].URL),
				},
				profile{
					Name:   p.Name + " (objects allocated)",
					URL:    baseURL + p.URL,
					Params: "?alloc_objects&url=" + url.QueryEscape(prof[i].URL),
				},
				profile{
					Name:   p.Name + " (space allocated)",
					URL:    baseURL + p.URL,
					Params: "?alloc_space&url=" + url.QueryEscape(prof[i].URL),
				},
			)
		} else {
			prof[i].URL = baseURL + p.URL
			prof[i].Params = "?url=" + url.QueryEscape(prof[i].URL)
		}
	}

	return
}

func (h *Handler) client() *http.Client {
	if h.Client != nil {
		return h.Client
	}
	return http.DefaultClient
}

func (h *Handler) prefix() string {
	if len(h.Prefix) != 0 {
		return h.Prefix
	}
	return "/debug/pprof/"
}

type service struct {
	Name     string    `json:"name"`
	Href     string    `json:"href"`
	Profiles []profile `json:"profiles,omitempty"`
}

type profile struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Params string `json:"params"`
}

func render(res http.ResponseWriter, req *http.Request, tpl *template.Template, val interface{}) {
	switch accept := strings.TrimSpace(req.Header.Get("Accept")); {
	case strings.Contains(accept, "text/html"):
		renderHTML(res, tpl, val)
	default:
		renderJSON(res, val)
	}
}

func renderJSON(res http.ResponseWriter, val interface{}) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewPrettyEncoder(res).Encode(val)
}

func renderHTML(res http.ResponseWriter, tpl *template.Template, val interface{}) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.Execute(res, val)
}

func splitPathQuery(s string) (path string, query string) {
	if i := strings.IndexByte(s, '?'); i >= 0 {
		path, query = s[:i], s[i+1:]
	} else {
		path = s
	}
	return
}
