package pprofserver

import (
	"context"
	"net"
)

type Service struct {
	Name  string
	Hosts []Host
}

type Host struct {
	Addr net.Addr
	Tags []string
}

type Registry interface {
	ListServices(ctx context.Context) ([]string, error)
	LookupService(ctx context.Context, name string) (Service, error)
}
