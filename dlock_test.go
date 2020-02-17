package test

import (
	"testing"
	"context"
	//"fmt"
	"time"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

var endpoints []string = []string{"127.0.0.1:2379"}

var lockTimeout  = 5 * time.Second

type lockerMutex struct{ *concurrency.Mutex }

func getLockPath(path string) string{
	return "/"+path
}


func NewLocker(session *concurrency.Session, pfx string) *lockerMutex {
	lockKey := getLockPath(pfx)
	return &lockerMutex{concurrency.NewMutex(session, lockKey)}
}

func (lm *lockerMutex) Lock() error {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	if err := lm.Mutex.Lock(ctx); err != nil {
		return err
	}
	return nil
}

func (lm *lockerMutex) Unlock() error {
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	return lm.Mutex.Unlock(ctx)
}


func TestLock(t *testing.T) {
	cfg1 := clientv3.Config{
		Endpoints: endpoints,
	}
	var err error
	client1, err := clientv3.New(cfg1)
	if err != nil {
		t.Fatalf("clientv3.New:%v", err)
	}
	session1, err := concurrency.NewSession(client1)
	if err != nil {
		t.Fatalf("concurrency.NewSession:%v", err)
	}

	locker1 := NewLocker(session1, "/locker/test")
	err = locker1.Lock()
	if err != nil {
		t.Fatalf("Lock error %v", err)
	}

	cfg2 := clientv3.Config{
		Endpoints: endpoints,
	}
	client2, err := clientv3.New(cfg2)
	if err != nil {
		t.Fatalf("clientv3.New:%v", err)
	}
	session2, err := concurrency.NewSession(client2)
	if err != nil {
		t.Fatalf("concurrency.NewSession:%v", err)
	}
	locker2 := NewLocker(session2, "/locker/test")
	err = locker2.Lock()
	if err == nil {
		t.Fatal("Lock does not take effect")
	}

	locker1.Unlock()
	err = locker2.Lock()
	if err != nil {
		t.Fatal("Unlock does not take effect")
	}
	locker2.Unlock()
}

