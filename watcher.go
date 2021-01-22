package main

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// Watch the latest revision
type EtcdWatcher struct {
	LastRevision int64
	Cancel       context.CancelFunc
	eventChan    chan *clientv3.Event
	Lock         *sync.RWMutex
	WatchKey     string

	incipientKVs []*mvccpb.KeyValue
	Client       *EtcdV3Client
}

//TODO: fix ctx
func (cl *EtcdV3Client) CreateEtcdWatcher(ctx context.Context, watch_prefix string) (*EtcdWatcher, error) {
	watcher := &EtcdWatcher{
		eventChan: make(chan *clientv3.Event, 100),
		Client:    cl,
		WatchKey:  watch_prefix,
		Lock:      new(sync.RWMutex),
	}

	//found revision by prefix
	resp, err := cl.Client.Get(ctx, watch_prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	watcher.LastRevision = resp.Header.Revision
	watcher.incipientKVs = resp.Kvs

	go func() {
		defer func() {
			var ret error
			if err := recover(); err != nil {
				if _, ok := err.(error); ok {
					ret = err.(error)
				} else {
					ret = fmt.Errorf("%+v", err)
				}
				cl.Logger.Error("[CreateEtcdWatcher]recover", zap.Any("err", ret))
			}
		}()

		//fork from cl.context or context.Background()
		ctx, cancel := context.WithCancel(context.Background())
		watcher.Cancel = cancel

		rch := cl.Client.Watch(ctx, watch_prefix, clientv3.WithPrefix(), clientv3.WithCreatedNotify(), clientv3.WithRev(watcher.LastRevision))
		for {
			for n := range rch {
				if n.CompactRevision > watcher.LastRevision {
					//update LastRevision
					watcher.LastRevision = n.CompactRevision
				}
				if n.Header.GetRevision() > watcher.LastRevision {
					//update LastRevision
					watcher.LastRevision = n.Header.GetRevision()
				}
				if err := n.Err(); err != nil {
					cl.Logger.Error("[CreateEtcdWatcher]Watch error", zap.Any("errmsg", err), zap.String("prefix", watch_prefix))
					continue
				}
				for _, ev := range n.Events {
					select {
					case watcher.eventChan <- ev:
						//notify a event
						cl.Logger.Info("[CreateEtcdWatcher]report up a event to eventChan", zap.Any("event", ev), zap.String("prefix", watch_prefix))
					default:
						cl.Logger.Error("[CreateEtcdWatcher]watch etcd with prefix err:block event chan, drop event message")
					}
				}
			}

			//recv a cancel,recreate a new one
			ctx, cancel := context.WithCancel(context.Background())
			watcher.Cancel = cancel
			if watcher.LastRevision > 0 {
				rch = cl.Client.Watch(ctx, watch_prefix, clientv3.WithPrefix(), clientv3.WithCreatedNotify(), clientv3.WithRev(watcher.LastRevision))
			} else {
				rch = cl.Client.Watch(ctx, watch_prefix, clientv3.WithPrefix(), clientv3.WithCreatedNotify())
			}
		}
	}()

	return watcher, nil
}

// 获取事件通知channel
func (w *EtcdWatcher) GetEventChan() chan *clientv3.Event {
	return w.eventChan
}

func (w *EtcdWatcher) GetIncipientKeyValues() []*mvccpb.KeyValue {
	return w.incipientKVs
}

// close 监听
func (w *EtcdWatcher) CloseWather() error {
	if w.Cancel != nil {
		w.Cancel()
	}
	return nil
}
