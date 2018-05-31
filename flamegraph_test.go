package pprofserver

import "testing"

func TestSupportsFlamegraph(t *testing.T) {
	url := "https://pprof.segment.build/tree?url=http%3A%2F%2F10.30.81.218%3A10240%2Fdebug%2Fpprof%2Fgoroutine%3Fdebug%3D1"
	if ok := supportsFlamegraph(url); !ok {
		t.Errorf("url=%s: want true, got %v", url, ok)
	}
	url = "https://pprof.segment.build/tree?url=http%3A%2F%2F10.30.81.218%3A10240%2Fdebug%2Fpprof%2Fgoroutine%3Fdebug%3D2"
	if ok := supportsFlamegraph(url); ok {
		t.Errorf("url=%s: want false, got %v", url, ok)
	}
}
