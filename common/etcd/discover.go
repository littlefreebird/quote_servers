package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"time"
)

type ServiceDiscover struct {
	cli        *clientv3.Client
	serverList map[string]string
	lock       sync.Mutex
}

type ServiceHandler func(key, val string) error

func NewServiceDiscover(endpoints []string) (*ServiceDiscover, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	sd := ServiceDiscover{
		cli:        cli,
		serverList: make(map[string]string),
	}
	return &sd, nil
}

func (sd *ServiceDiscover) GetServices() map[string]string {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	return sd.serverList
}

func (sd *ServiceDiscover) DeleteService(key string) {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	delete(sd.serverList, key)
}

func (sd *ServiceDiscover) SetService(key, val string) {
	sd.lock.Lock()
	defer sd.lock.Unlock()
	sd.serverList[key] = val
}

func (sd *ServiceDiscover) Close() error {
	return sd.cli.Close()
}

func (sd *ServiceDiscover) WatchService(prefix string, putFunc, delFunc ServiceHandler) error {
	rsp, err := sd.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, item := range rsp.Kvs {
		sd.SetService(string(item.Key), string(item.Value))
		if putFunc != nil {
			putFunc(string(item.Key), string(item.Value))
		}
	}
	go sd.watcher(prefix, putFunc, delFunc)
	return nil
}

func (sd *ServiceDiscover) watcher(prefix string, putFunc, delFunc ServiceHandler) {
	ch := sd.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for item := range ch {
		for _, ev := range item.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				sd.SetService(string(ev.Kv.Key), string(ev.Kv.Value))
				if putFunc != nil {
					putFunc(string(ev.Kv.Key), string(ev.Kv.Value))
				}
			case clientv3.EventTypeDelete:
				sd.DeleteService(string(ev.Kv.Key))
				if delFunc != nil {
					delFunc(string(ev.Kv.Key), string(ev.Kv.Value))
				}
			}
		}
	}
}

func (sd *ServiceDiscover) GetServicesWithPrefix(prefix string) (map[string]string, error) {
	rsp, err := sd.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	ret := make(map[string]string)
	for _, item := range rsp.Kvs {
		ret[string(item.Key)] = string(item.Value)
	}
	return ret, nil
}
