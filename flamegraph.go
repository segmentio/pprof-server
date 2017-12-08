package pprofserver

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/uber/go-torch/pprof"
	"github.com/uber/go-torch/renderer"
)

func renderFlamegraph(w io.Writer, url string, pprofArgs []string) error {
	c := exec.Command("go", "tool", "pprof", "-raw", url)
	raw, err := c.Output()
	if err != nil {
		return fmt.Errorf("get raw pprof data: %v", err)
	}

	profile, err := pprof.ParseRaw(raw)
	if err != nil {
		return fmt.Errorf("parse raw pprof output: %v", err)
	}

	sampleIndex := pprof.SelectSample(pprofArgs, profile.SampleNames)
	flameInput, err := renderer.ToFlameInput(profile, sampleIndex)
	if err != nil {
		return fmt.Errorf("convert stacks to flamegraph input: %v", err)
	}

	title := url
	if len(pprofArgs) > 0 {
		title = fmt.Sprintf("%s (%s)", title, strings.Join(pprofArgs, " "))
	}
	flameGraph, err := renderer.GenerateFlameGraph(flameInput, "--title", title)
	if err != nil {
		return fmt.Errorf("generate flame graph: %v", err)
	}

	if _, err := w.Write(flameGraph); err != nil {
		return fmt.Errorf("write flamegraph SVG: %v", err)
	}

	return nil
}
