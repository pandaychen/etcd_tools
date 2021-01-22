package etcd_tools

import (
	"time"

	zaplog "github.com/pandaychen/goes-wrapper/zaplog"
	"go.uber.org/zap"
)

type EtcdConfig struct {
	Endpoints      []string      `json:"endpoints"`
	ConnectTimeout time.Duration `json:"timeout"`
	Secure         bool          `json:"secure"`
	TTL            int           // 单位：s
	Logger         *zap.Logger

	//Etcd
	DialKeepAliveTime    time.Duration `json:"dialkeepalivetime"`
	DialKeepAliveTimeout time.Duration `json:"dialkeepalivetimeout"`

	//ETCD 认证参数
	CertFilePath string `json:"certfilepath"`
	KeyFilePath  string `json:"keyfilepath"`
	CaCertPath   string `json:"cacertpath"`
	BasicAuth    bool   `json:"basicauth"`
	UserName     string `json:"username"`
	Password     string `json:"passwd"`
}

func DefaultConfig() *EtcdConfig {
	logger, _ := zaplog.ZapLoggerInit("etcdv3-client")
	return &EtcdConfig{
		Endpoints:      []string{"127.0.0.1:2379"},
		BasicAuth:      false,
		ConnectTimeout: time.Duration(6 * time.Second),
		Secure:         false,
		Logger:         logger,
	}
}
