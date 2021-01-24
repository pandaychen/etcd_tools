package etcd_tools

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	zaplog "github.com/pandaychen/goes-wrapper/zaplog"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// a EtcdV3Client represent a etcdv3 client
type EtcdV3Client struct {
	Etcdconf *EtcdConfig
	Context  context.Context
	Cancel   context.CancelFunc
	Logger   *zap.Logger

	*clientv3.Client
}

func NewEtcdV3Client(config *EtcdConfig) (*EtcdV3Client, error) {
	//init etcdv3 client config
	etcdconf := clientv3.Config{
		Endpoints:            config.Endpoints,
		DialTimeout:          config.ConnectTimeout,
		DialKeepAliveTime:    config.DialKeepAliveTime,
		DialKeepAliveTimeout: config.DialKeepAliveTimeout,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
	}
	var tlsEnabled bool = false
	ctx, cancel := context.WithCancel(context.Background())

	etcdclient := &EtcdV3Client{
		Etcdconf: config,
		Logger:   config.Logger,
		Context:  ctx,
		Cancel:   cancel,
	}

	if etcdclient.Logger == nil {
		etcdclient.Logger, _ = zaplog.ZapLoggerInit("etcdv3-client")
	}

	if config.Endpoints == nil || len(config.Endpoints) == 0 {
		return nil, errors.New("Must specify etcd endpoints")
	}

	if !config.Secure {
		etcdconf.DialOptions = append(etcdconf.DialOptions, grpc.WithInsecure())
	}

	if config.BasicAuth {
		if config.UserName == "" || config.Password == "" {
			return nil, errors.New("Must specify etcd basic auth username or password")
		}
		etcdconf.Username = config.UserName
		etcdconf.Password = config.Password
	}

	//check if connect with tls
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if config.CaCertPath != "" {
		certBytes, err := ioutil.ReadFile(config.CaCertPath)
		if err != nil {
			etcdclient.Logger.Error("[newEtcdV3Client]parse CaCert error", zap.String("errmsg", err.Error()))
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
		tlsEnabled = true
	}

	if config.CertFilePath != "" && config.KeyFilePath != "" {
		tlsCert, err := tls.LoadX509KeyPair(config.CertFilePath, config.KeyFilePath)
		if err != nil {
			etcdclient.Logger.Error("[newEtcdV3Client]parse CertFile or KeyFile error", zap.String("errmsg", err.Error()))
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
		tlsEnabled = true
	}

	if tlsEnabled {
		etcdconf.TLS = tlsConfig
	}

	client, err := clientv3.New(etcdconf)

	if err != nil {
		etcdclient.Logger.Error("[newEtcdV3Client]create  etcd client error", zap.String("errmsg", err.Error()))
		return nil, err
	}

	etcdclient.Client = client
	return etcdclient, nil
}
