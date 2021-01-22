package etcd_tools

import (
	"context"
	"errors"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"go.uber.org/zap"
)

func (cl *EtcdV3Client) GetKeyValueResision(ctx context.Context, key string) (string, int64, error) {
	resp, err := cl.Client.Get(ctx, key)
	if err != nil {
		cl.Logger.Error("[GetKeyValueResision]error", zap.String("errmsg", err.Error()))
		return "", -1, err
	}

	if len(resp.Kvs) > 0 {

		return string(resp.Kvs[0].Value), resp.Header.Revision, nil
	}

	return "", -1, errors.New("no value")
}

func (cl *EtcdV3Client) GetKeyValue(ctx context.Context, key string) (*mvccpb.KeyValue, error) {
	resp, err := cl.Client.Get(ctx, key)
	if err != nil {
		cl.Logger.Error("[GetKeyValue]error", zap.String("errmsg", err.Error()))
		return "", err
	}

	if len(resp.Kvs) > 0 {
		return rp.Kvs[0], nil
	}

	return "", errors.New("no value")
}

func (cl *EtcdV3Client) GetKeyPrefixValues(ctx context.Context, key_prefix string) (map[string]string, error) {
	var values = make(map[string]string)

	resp, err := cl.Client.Get(ctx, key_prefix, clientv3.WithPrefix())
	if err != nil {
		cl.Logger.Error("[GetKeyPrefixValues]error", zap.String("prefix", key_prefix), zap.String("errmsg", err.Error()))
		return nil, err
	}

	for _, kv := range resp.Kvs {
		values[string(kv.Key)] = string(kv.Value)
	}

	return values, nil
}

func (cl *EtcdV3Client) DelKey(ctx context.Context, key string) (int64, error) {
	resp, err := cl.Client.Delete(ctx, key)
	if err != nil {
		cl.Logger.Error("[DelKey]error", zap.String("key", key), zap.String("errmsg", err.Error()))
		return -1, err
	}
	return resp.Deleted, err
}

func (cl *EtcdV3Client) DelKeyPrefix(ctx context.Context, key_prefix string) (deleted int64, err error) {
	resp, err := client.Delete(ctx, key_prefix, clientv3.WithPrefix())
	if err != nil {
		cl.Logger.Error("[DelKeyPrefix]error", zap.String("prefix", key_prefix), zap.String("errmsg", err.Error()))
		return 0, err
	}
	return resp.Deleted, err
}
