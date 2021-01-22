package etcd_tools

import (
	"context"
	"errors"

	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

// 封装一个可以自动续期的key值
type EtcdLeaseKey struct {
	Key      string
	Value    string
	Client   *EtcdV3Client
	LeaseTtl int64 //ETCD-lease
	Leaseid  clientv3.LeaseID

	KeepAliveSign bool

	LeaseRespChan <-chan *clientv3.LeaseKeepAliveResponse

	// cancel
	Context context.Context
	Cancel  context.CancelFunc
}

func CreateEtcdLeaseKey(etcdclient *EtcdV3Client, lease_ttl int64, keepalive_sign bool, key, value string) (*EtcdLeaseKey, error) {
	if etcdclient == nil {
		return nil, errors.New("param illegal")
	}

	leasekey := EtcdLeaseKey{
		Client:        etcdclient,
		LeaseTtl:      lease_ttl,
		KeepAliveSign: keepalive_sign,
		Key:           key,
		Value:         value,
	}
	if leasekey.LeaseTtl > int64(0) {
		resp, err := leasekey.Client.Client.Grant(leasekey.Client.Context, int64(leasekey.LeaseTtl))
		if err != nil {
			leasekey.Client.Logger.Error("[CreateEtcdLeaseKey]create clientv3 lease failed", zap.String("errmsg", err.Error()))
			return nil, err
		}
		leasekey.Leaseid = resp.ID
		ctx, cancel := context.WithCancel(leasekey.Client.Context)
		if _, err := leasekey.Client.Client.Put(ctx, leasekey.Key, leasekey.Value, clientv3.WithLease(leasekey.Leaseid)); err != nil {
			leasekey.Client.Logger.Error("[CreateEtcdLeaseKey]Set service with ttl failed", zap.String("errmsg", err.Error()), zap.String("key", leasekey.Key))
			return nil, err
		}

		leasekey.Context = ctx
		leasekey.Cancel = cancel

		//keepalive
		if leasekey.KeepAliveSign {
			//in keepalive,start with a new groutine for loop
			leaseRespChan, err := leasekey.Client.Client.KeepAlive(ctx, leasekey.Leaseid)
			if err != nil {
				leasekey.Client.Logger.Error("[CreateEtcdLeaseKey]KeepAlive with ttl failed", zap.String("errmsg", err.Error()), zap.String("key", leasekey.Key))
				return nil, err
			}
			leasekey.LeaseRespChan = leaseRespChan
			go leasekey.ListenLeaseChan()

		}

	}

	return &leasekey, nil
}

func (lk *EtcdLeaseKey) ListenLeaseChan() {
	var (
		leaseKeepResp *clientv3.LeaseKeepAliveResponse
	)
	for {
		select {
		case leaseKeepResp = <-lk.LeaseRespChan:
			if leaseKeepResp == nil {
				lk.Client.Logger.Error("[ListenLeaseChan]Etcd leaseid not effectiveness,quit", zap.String("key", lk.Key), zap.Int64("leaseid", int64(lk.Leaseid)))
				//TODO:need alarm
				goto END
			} else {
				lk.Client.Logger.Info("[ListenLeaseChan]Etcd leaseid effectiveness", zap.String("key", lk.Key), zap.Int64("leaseid", int64(leaseKeepResp.ID)))
			}
		}
	}
END:
}

func (lk *EtcdLeaseKey) CancelLease() {
	// wait deregister then delete
	lk.Client.Logger.Info("[CancelLease]Etcd register quit(With auto Keepalive)")
	defer lk.Cancel()
	if lk.Client != nil {
		count, err := lk.Client.DelKey(lk.Context, lk.Key)
		if err != nil {
			lk.Client.Logger.Info("[CancelLease]Delete key error", zap.String("key", lk.Key), zap.String("errmsg", err.Error()), zap.Int64("delcount", count))
			return
		}

	}

	return
}
