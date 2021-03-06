package pprofserver

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"path"
	"strings"

	"github.com/segmentio/events"
	"github.com/uber/go-torch/pprof"
	"github.com/uber/go-torch/renderer"
)

func supportsFlamegraph(params string) bool {
	if strings.HasPrefix(params, "?") {
		params = params[1:]
	}
	query, err := url.ParseQuery(params)
	if err != nil {
		events.Log("flamegraph support check: params=%{params}q: %{error}s", params, err)
		return false
	}
	pprofUrl := query.Get("url")
	pprofParsed, err := url.Parse(pprofUrl)
	if err != nil {
		return false
	}
	switch path.Base(pprofParsed.Path) {
	case "profile", "heap", "block", "mutex":
		return true
	case "goroutine":
		return pprofParsed.Query().Get("debug") == "1"
	}
	return false
}

func renderFlamegraph(w io.Writer, url, sampleType string) error {
	// Get the raw pprof data
	c := exec.Command("go", "tool", "pprof", "-raw", url)
	raw, err := c.Output()
	if err != nil {
		return fmt.Errorf("get raw pprof data: %v", err)
	}

	profile, err := pprof.ParseRaw(raw)
	if err != nil {
		return fmt.Errorf("parse raw pprof output: %v", err)
	}

	// Select a sample type from the profile (bytes allocated, objects allocated, etc.)
	var args []string
	if sampleType != "" {
		args = append(args, "-"+sampleType)
	}
	sampleIndex := pprof.SelectSample(args, profile.SampleNames)
	flameInput, err := renderer.ToFlameInput(profile, sampleIndex)
	if err != nil {
		return fmt.Errorf("convert stacks to flamegraph input: %v", err)
	}

	// Construct graph title
	title := url
	if sampleType != "" {
		title = fmt.Sprintf("%s (%s)", url, sampleType)
	}

	// Try to find reasonable units
	unit := "samples"
	if strings.Contains(sampleType, "space") {
		unit = "bytes"
	} else if strings.Contains(sampleType, "objects") {
		unit = "objects"
	}

	// Render the graph
	flameGraph, err := renderer.GenerateFlameGraph(flameInput, "--title", title, "--countname", unit)
	if err != nil {
		return fmt.Errorf("generate flame graph: %v", err)
	}

	// Write the graph to the response
	if _, err := w.Write(flameGraph); err != nil {
		return fmt.Errorf("write flamegraph SVG: %v", err)
	}

	return nil
}
