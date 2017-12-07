package pprofserver

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/uber/go-torch/pprof"
	"github.com/uber/go-torch/renderer"
)

func renderFlamegraph(w io.Writer, r io.Reader) error {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read pprof output: %v", err)
	}

	profile, err := pprof.ParseRaw(raw)
	if err != nil {
		return fmt.Errorf("parse pprof output: %v", err)
	}

	sampleIndex := pprof.SelectSample(nil, profile.SampleNames)
	flameInput, err := renderer.ToFlameInput(profile, sampleIndex)
	if err != nil {
		return fmt.Errorf("convert stacks to flamegraph input: %v", err)
	}

	flameGraph, err := renderer.GenerateFlameGraph(flameInput)
	if err != nil {
		return fmt.Errorf("generate flame graph: %v", err)
	}

	if _, err := w.Write(flameGraph); err != nil {
		return fmt.Errorf("write flamegraph SVG: %v", err)
	}

	return nil
}
