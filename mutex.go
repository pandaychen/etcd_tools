package etcd_tools

import (
	"context"
	"time"

	"go.etcd.io/etcd/clientv3/concurrency"
	"go.uber.org/zap"
)

//分布式Mutex锁

type EtcdMutex struct {
	LockKey string
	Client  *EtcdV3Client
	Session *concurrency.Session
	Mutex   *concurrency.Mutex
	Timeout time.Duration
}

func (cl *EtcdV3Client) CreateEtcdMutex(key string, timeout time.Duration) (*EtcdMutex, error) {
	mutex := &EtcdMutex{
		LockKey: key,
		Client:  cl,
		Timeout: timeout}

	var err error

	var opts []concurrency.SessionOption
	if timeout != time.Duration(0) {
		opts = append(opts, concurrency.WithTTL(int(timeout)))
	}

	mutex.Session, err = concurrency.NewSession(cl.Client, opts...)
	if err != nil {
		m.Client.Logger.Error("[UnloCreateEtcdMutexck]NewSession error", zap.String("errmsg", err.Error()), zap.String("key", key))
		return nil, err
	}
	mutex.Mutex = concurrency.NewMutex(mutex.Session, key)
	return mutex, nil
}

//Lock
func (m *EtcdMutex) Lock() error {
	ctx, _ := context.WithCancel(m.Client.Context)

	return m.Mutex.Lock(ctx)
}

func (m *EtcdMutex) TryLock(timeout time.Duration) error {
	gctx, cancel := context.WithTimeout(m.Client.Context, timeout)
	defer cancel()
	return m.Mutex.Lock(gctx)
}

// Unlock
func (m *EtcdMutex) Unlock() (err error) {
	err = m.Mutex.Unlock(m.Client.Context)
	if err != nil {
		m.Client.Logger.Error("[Unlock]error", zap.String("errmsg", err.Error()))
		return nil
	}
	return m.Session.Close()
}
