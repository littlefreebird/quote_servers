package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type QuoteLock struct {
	Locker *concurrency.Mutex
	Name   string
	TTL    int
}

func CreateLocker(addr string, ttl int, name string) (*QuoteLock, error) {
	var ql QuoteLock
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{addr},
	})
	if err != nil {
		return nil, err
	}
	session, err := concurrency.NewSession(client, concurrency.WithTTL(ttl))
	ql.Locker = concurrency.NewMutex(session, name)
	ql.Name = name
	ql.TTL = ttl
	return &ql, nil
}

func (l *QuoteLock) Lock() {
	l.Locker.Lock(context.TODO())
}

func (l *QuoteLock) Unlock() {
	l.Locker.Unlock(context.TODO())
}
