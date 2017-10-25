package pprofserver

import (
	"context"
	"sort"

	consul "github.com/segmentio/consul-go"
)

type ConsulRegistry struct{}

func (r *ConsulRegistry) ListServices(ctx context.Context) ([]string, error) {
	services, err := consul.ListServices(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0, len(services))

	for srv := range services {
		list = append(list, srv)
	}

	sort.Strings(list)
	return list, nil
}

func (r *ConsulRegistry) LookupService(ctx context.Context, name string) (Service, error) {
	svc := Service{}

	endpoints, err := consul.LookupService(ctx, name)
	if err != nil {
		return svc, err
	}

	svc.Hosts = make([]Host, 0, len(endpoints))
	for _, endpoint := range endpoints {
		svc.Hosts = append(svc.Hosts, Host{
			Addr: endpoint.Addr,
			Tags: endpoint.Tags,
		})
	}
	return svc, nil
}
