// Copyright (c) 2015-2022 CloudJ Technology Co., Ltd.

package common

import (
	"cloudiac/configs"
	"cloudiac/utils/consul"
	"cloudiac/utils/logs"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
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
func CheckAndReConnectConsul(serviceName string) error {
	lg := logs.Get().WithField("func", "CheckAndReConnectConsul")
	// 首次启动获取锁
	if err := start(serviceName, true); err != nil {
		lg.Warnf("start failed, error: %v", err)
		return err
	}

	// 锁丢失后重新获取锁，获取之后重新注册服务
	go func() {
		for {
			err := start(serviceName, false)
			if err != nil {
				lg.Warnf("restart failed, error: %v", err)
			}
			time.Sleep(time.Second * 10)
		}
	}()

	return nil
}

func start(serviceName string, isTryOnce bool) error {
	lg := logs.Get().WithField("func", "CheckAndReConnectConsul->start")
	ctx := context.Background()
	defer ctx.Done()

	// 从配置文件中获取当前服务的ID
	conf := configs.Get()
	serviceId := conf.Consul.ServiceID

	lg.Infof("acquire %s lock ...", serviceId)

	var err error
	var lockLostCh = make(<-chan struct{})
	for {
		lockLostCh, err = acquireLock(ctx, serviceId, isTryOnce)
		if err == nil {
			break
		}
		if err != nil && isTryOnce {
			return err
		}

		// 正常情况下 acquireLock 会阻塞直到成功获取锁，如果报错了就是出现了异常(可能是连接问题)
		lg.Errorf("acquire %s lock failed: %v", serviceId, err)
		time.Sleep(time.Second * 10)
	}

	// 注册服务
	err = ServiceRegister(serviceName)
	if isTryOnce {
		return err
	}

	if err != nil {
		return err
	}

	// 丢失lock时，中止
	<-lockLostCh
	lg.Warnf("disconnected from consul")
	return nil
}

func acquireLock(ctx context.Context, serviceId string, isTryOnce bool) (<-chan struct{}, error) {
	locker, err := consul.GetLocker(serviceId+"-lock", []byte(serviceId), configs.Get().Consul.Address, isTryOnce)
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
	if lockLostCh == nil {
		return nil, errors.Wrap(fmt.Errorf("lock is already held by others"), "acquire lock")
	}
	lockHeld = true
	return lockLostCh, nil
}
