// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package common

import (
	"cloudiac/configs"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	IacPortalLockKey = "iac-portal-lock"
)

func ServiceRegister(serviceName string) error {
	conf := configs.Get()
	logger := logs.Get()

	logger.Debugf("Start register %s service", serviceName)
	err := consul.Register(serviceName, conf.Consul)
	if err != nil {
		logger.Debug("Service register failied: %s", err)
	} else {
		logger.Debug("Service register success.")
	}

	return err
}

func ReRegisterService(register bool, serviceName string) error {
	err := ServiceRegister(serviceName)
	if err != nil {
		return err
	}

	if register {
		os.Exit(0)
	}
	return nil
}

// 丢失consul连接时，尝试重新连接
func CheckAndReConnectConsul(serviceName string) {
	start()

	for {
		start()
		time.Sleep(time.Second * 10)
	}

	//ServiceRegister(serviceName)
}

func start() {
	lg := logs.Get().WithField("func", "CheckAndReConnectConsul->start")
	ctx := context.Background()
	defer ctx.Done()

	lg.Infof("acquire iac portal lock ...")
	var err error
	var lockLostCh = make(<-chan struct{})
	for {
		lockLostCh, err = acquireLock(ctx)
		if err == nil {
			break
		}

		// 正常情况下 acquireLock 会阻塞直到成功获取锁，如果报错了就是出现了异常(可能是连接问题)
		lg.Errorf("acquire iac portal lock failed: %v", err)
		time.Sleep(time.Second * 10)
	}

	// 丢失lock时，中止
	<-lockLostCh
}

func acquireLock(ctx context.Context) (<-chan struct{}, error) {
	locker, err := consul.GetLocker(IacPortalLockKey, []byte(IacPortalLockKey), configs.Get().Consul.Address)
	if err != nil {
		return nil, errors.Wrap(err, "get locker")
	}

	stopLockCh := make(chan struct{})
	lockHeld := false
	go func() {
		<-ctx.Done()
		close(stopLockCh)
		if lockHeld {
			if err := locker.Unlock(); err != nil {
				logs.Get().Errorf("release lock error: %v", err)
			}
		}
	}()

	lockLostCh, err := locker.Lock(stopLockCh)
	if err != nil {
		return nil, errors.Wrap(err, "acquire lock")
	}
	lockHeld = true
	return lockLostCh, nil
}
