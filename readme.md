## EtcdV3 库的接口封装

- 客户端创建
- 常用操作
- Lease 租约
- Mutex：分布式锁
- 监听（前缀）器

## 使用示例

```golang
import (
        "context"
        "fmt"
        etcdlib "github.com/pandaychen/etcd_tools"
        "go.etcd.io/etcd/mvcc/mvccpb"
        "time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cli, _ := etcdlib.NewEtcdV3Client(etcdlib.DefaultConfig())
	go func() {
			time.Sleep(3 * time.Second)
			fmt.Println(cli.GetKeyValueResision(ctx, "/test/key"))
			fmt.Println(cli.DelKey(ctx, "/test/key"))
			fmt.Println(cli.GetKeyValueResision(ctx, "/test/key"))
	}()

	//lease
	l, _ := etcdlib.CreateEtcdLeaseKey(cli, 100, true, "test", "value")
	l.CancelLease()
	w, _ := cli.CreateEtcdWatcher(ctx, "/test/")

	//watcher
	for event := range w.GetEventChan() {
			switch event.Type {
			case mvccpb.PUT:
					fmt.Println("mvccpb.PUT")
			case mvccpb.DELETE:
					fmt.Println("mvccpb.DELETE")
			}

	}
}
```
