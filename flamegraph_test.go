package pprofserver

import "testing"

func TestSupportsFlamegraph(t *testing.T) {
	params := "?url=http%3A%2F%2F10.30.81.218%3A10240%2Fdebug%2Fpprof%2Fgoroutine%3Fdebug%3D1"
	if ok := supportsFlamegraph(params); !ok {
		t.Errorf("params=%s: want true, got %v", params, ok)
	}
	params = "?url=http%3A%2F%2F10.30.81.218%3A10240%2Fdebug%2Fpprof%2Fgoroutine%3Fdebug%3D2"
	if ok := supportsFlamegraph(params); ok {
		t.Errorf("url=%s: want false, got %v", params, ok)
	}
}
