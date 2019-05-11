package main

import (
	"flag"
	"github.com/pjoc-team/base-service/pkg/service"
	"github.com/pjoc-team/pay-gateway/pkg/gateway"
	"github.com/pjoc-team/pay-gateway/pkg/rest"
)

var (
	listenAddr            = flag.String("listen-addr", ":8080", "HTTP listen address.")
	configURI             = flag.String("c", "config.yaml", "uri to load config")
	tlsEnable             = flag.Bool("tls", false, "enable tls")
	logLevel              = flag.String("log-level", "debug", "logger level")
	logFormat             = flag.String("log-format", "text", "text or json")
	caCert                = flag.String("ca-cert", service.WithConfigDir("ca.pem"), "Trusted CA certificate.")
	tlsCert               = flag.String("tls-cert", service.WithConfigDir("cert.pem"), "TLS server certificate.")
	tlsKey                = flag.String("tls-key", service.WithConfigDir("key.pem"), "TLS server key.")
	serviceName           = flag.String("s", "", "PayGatewayService name in service discovery.")
	registerServiceToEtcd = flag.Bool("r", true, "Register service to etcd.")
	etcdPeers             = flag.String("etcd-peers", "", "Etcd peers. example: 127.0.0.1:2379,127.0.0.1:12379")
)

func main() {
	flag.Parse()
	//serviceDir := gateway.ETCD_DIR_ROOT + "/services"
	svc := service.InitService(*listenAddr,
		*configURI,
		*tlsEnable,
		*logLevel,
		*logFormat,
		*caCert,
		*tlsCert,
		*tlsKey,
		*serviceName,
		*registerServiceToEtcd,
		*etcdPeers,
		"")
	go rest.Start()

	gateway.Init(svc)
}
