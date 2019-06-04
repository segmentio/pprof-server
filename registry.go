package pprofserver

import (
	"context"
	"fmt"
	"net"
)

type Service struct {
	Name  string
	Hosts []Host
}

func (s Service) String() string {
	return "service"
}

func (s Service) ListServices(ctx context.Context) ([]string, error) {
	return []string{s.Name}, nil
}

func (s Service) LookupService(ctx context.Context, name string) (Service, error) {
	if s.Name == name {
		return s, nil
	}
	return Service{}, fmt.Errorf("%s: service not found", name)
}

type Host struct {
	Addr net.Addr
	Tags []string
}

type Registry interface {
	ListServices(ctx context.Context) ([]string, error)
	LookupService(ctx context.Context, name string) (Service, error)
	String() string
}
