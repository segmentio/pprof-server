package main

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/segmentio/conf"
	consul "github.com/segmentio/consul-go"
	"github.com/segmentio/events"
	_ "github.com/segmentio/events/ecslogs"
	"github.com/segmentio/events/httpevents"
	_ "github.com/segmentio/events/text"
	pprofserver "github.com/segmentio/pprof-server"
	"github.com/segmentio/stats"
	"github.com/segmentio/stats/datadog"
	"github.com/segmentio/stats/httpstats"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	type dogstatsdConfig struct {
		Address    string `conf:"address"     help:"Address of the dogstatsd server to send metrics to."`
		BufferSize int    `conf:"buffer-size" help:"Buffer size of the dogstatsd client." validet:"min=1024"`
	}

	type kubernetesConfig struct {
		MasterURL  string `conf:"master-url" help:"Master of the Kubernetes URL."`
		Kubeconfig string `conf:"kubeconfig" help:"Path of the Kubeconfig file."`
	}

	config := struct {
		Bind     string `conf:"bind"      help:"Network address to listen on." validate:"nonzero"`
		Registry string `conf:"registry"  help:"Address of the registry used to discover services."`
		Debug    bool   `conf:"debug"     help:"Enable debug mode."`

		Dogstatsd dogstatsdConfig `conf:"dogstatsd" help:"Configuration of the dogstatsd client."`

		Kubernetes kubernetesConfig `conf:"kubernetes" help:"Kubernetes configuration."`
	}{
		Bind: ":6061",
		Dogstatsd: dogstatsdConfig{
			BufferSize: 1024,
		},
	}
	conf.Load(&config)

	events.DefaultLogger.EnableDebug = config.Debug
	events.DefaultLogger.EnableSource = config.Debug
	defer stats.Flush()

	if len(config.Dogstatsd.Address) != 0 {
		stats.Register(datadog.NewClientWith(datadog.ClientConfig{
			Address:    config.Dogstatsd.Address,
			BufferSize: config.Dogstatsd.BufferSize,
		}))
	}

	ctx, cancel := events.WithSignals(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var registry pprofserver.Registry
	if len(config.Registry) != 0 {
		u, err := url.Parse(config.Registry)
		if err != nil {
			events.Log("bad registry URL: %{url}s", config.Registry)
			os.Exit(1)
		}
		switch u.Scheme {
		case "":
			events.Log("no registry is configured")
		case "consul":
			consul.DefaultClient.Address = u.Host
			consul.DefaultResolver.Balancer = nil
			registry = &pprofserver.ConsulRegistry{}
			events.Log("using consul registry at %{address}s", u.Host)
		case "service":
			name := strings.TrimPrefix(u.Path, "/")
			if name == "" {
				name = "Service"
			}
			registry = pprofserver.Service{
				Name:  name,
				Hosts: []pprofserver.Host{{Addr: hostAddr(u.Host)}},
			}
			events.Log("using single service registry at %{address}s", u.Host)
		case "kubernetes":
			kubeclient, err := kubeClient(
				u.Host == "in-cluster",
				config.Kubernetes.MasterURL,
				config.Kubernetes.Kubeconfig)
			if err != nil {
				panic(err)
			}

			kubernetesRegistry := pprofserver.NewKubernetesRegistry(kubeclient)
			kubernetesRegistry.Init(ctx)

			registry = kubernetesRegistry
			events.Log("using kubernetes registry")
		default:
			events.Log("unsupported registry: %{url}s", config.Registry)
			os.Exit(1)
		}
	}

	var httpTransport = http.DefaultTransport
	httpTransport = httpstats.NewTransport(httpTransport)
	httpTransport = httpevents.NewTransport(httpTransport)
	http.DefaultTransport = httpTransport

	var httpHandler http.Handler = &pprofserver.Handler{Registry: registry}
	httpHandler = httpstats.NewHandler(httpHandler)
	httpHandler = httpevents.NewHandler(httpHandler)
	http.Handle("/", httpHandler)

	server := http.Server{
		Addr: config.Bind,
	}

	go func() {
		<-ctx.Done()
		cancel()
		server.Shutdown(context.Background())
	}()

	events.Log("pprof server is listening for incoming connections on %{address}s", config.Bind)

	switch err := server.ListenAndServe(); {
	case err == http.ErrServerClosed:
	case events.IsTermination(err):
	case events.IsInterruption(err):
	default:
		events.Log("fatal error: %{error}s", err)
	}
}

type hostAddr string

func (a hostAddr) Network() string { return "" }
func (a hostAddr) String() string  { return string(a) }

func kubeClient(inCluster bool, master, kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if inCluster {
		events.Log("kubernetes: using pod service account with InClusterConfig.")
		config, err = rest.InClusterConfig()
	} else {
		events.Log("kubernetes: kubeconfig file.")
		config, err = clientcmd.BuildConfigFromFlags(master, kubeconfig)
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
