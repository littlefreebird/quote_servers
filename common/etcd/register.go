package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type ServiceRegister struct {
	cli          *clientv3.Client
	leaseID      clientv3.LeaseID
	keepLiveChan <-chan *clientv3.LeaseKeepAliveResponse
	key          string
	val          string
}

func NewServiceRegister(endpoints []string, key string, val string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	sr := ServiceRegister{
		cli: cli,
		key: key,
		val: val,
	}
	err = sr.putKeyWithLease(lease)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

func (r *ServiceRegister) putKeyWithLease(lease int64) error {
	rsp, err := r.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	_, err = r.cli.Put(context.Background(), r.key, r.val, clientv3.WithLease(rsp.ID))
	if err != nil {
		return err
	}
	r.keepLiveChan, err = r.cli.KeepAlive(context.Background(), rsp.ID)
	if err != nil {
		return err
	}
	r.leaseID = rsp.ID
	return nil
}

func (r *ServiceRegister) GetLeaseRspChan() <-chan *clientv3.LeaseKeepAliveResponse {
	return r.keepLiveChan
}

func (r *ServiceRegister) Close() error {
	if _, err := r.cli.Revoke(context.Background(), r.leaseID); err != nil {
		return err
	}
	return r.cli.Close()
}
