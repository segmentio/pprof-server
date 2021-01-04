module github.com/segmentio/pprof-server

go 1.15

require (
	github.com/fatih/color v1.7.0 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/segmentio/conf v0.0.0-20170612230246-5d701c9ec529
	github.com/segmentio/consul-go v0.0.0-20170912072050-42ff3637e5db
	github.com/segmentio/events v2.0.1+incompatible
	github.com/segmentio/fasthash v1.0.0 // indirect
	github.com/segmentio/objconv v0.0.0-20170810202704-5dca7cbec799
	github.com/segmentio/stats v0.0.0-20170908015358-6da51b6c447b
	github.com/uber/go-torch v0.0.0-20170825044957-ddbe52cdc30e
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	gopkg.in/validator.v2 v2.0.0-20170814132753-460c83432a98 // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v11.0.0+incompatible
)

replace k8s.io/client-go => k8s.io/client-go v0.20.1
