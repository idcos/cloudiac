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
func CheckAndReConnectConsul(serviceName string, serviceId string) error {
	lg := logs.Get().WithField("func", "CheckAndReConnectConsul")

	// 首次启动获取锁并注册服务
	lockLostCh, cancelCtx, err := lockAndRegister(serviceName, serviceId, true)
	if err != nil {
		lg.Warnf("start failed, error: %v", err)

		// 首次启动失败后， 等待 SessionTTL(默认10s) 时间后再次尝试获取锁
		sleepSeconds := configs.Get().Consul.WaitLockRelease
		if sleepSeconds == 0 {
			sleepSeconds = 20
		}
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
		lockLostCh, cancelCtx, err = lockAndRegister(serviceName, serviceId, true)
		if err != nil {
			lg.Warnf("the second time start failed, error: %v", err)
			return err
		}
	}

	go func() {
		for {
			// lock 丢失，主动结束 ctx 避免 context 泄漏
			<-lockLostCh
			cancelCtx()
			lg.Warnf("disconnected from consul")

			// 锁丢失后重新获取锁，并重新注册服务
			for {
				lockLostCh, cancelCtx, err = lockAndRegister(serviceName, serviceId, false)
				if err != nil {
					lg.Warnf("restart failed, error: %v", err)
					time.Sleep(time.Second * 10)
				} else {
					break
				}
			}
		}
	}()

	return nil
}

func lockAndRegister(serviceName string, serviceId string, isTryOnce bool) (<-chan struct{}, context.CancelFunc, error) {
	lg := logs.Get().WithField("func", "CheckAndReConnectConsul->start")
	lg.Infof("acquire %s lock ...", serviceId)

	ctx, cancel := context.WithCancel(context.Background())

	lockLostCh, err := acquireLock(ctx, serviceId, isTryOnce)
	if err != nil {
		// 正常情况下 acquireLock 会阻塞直到成功获取锁，如果报错了就是出现了异常(可能是连接问题)
		lg.Errorf("acquire %s lock failed: %v", serviceId, err)
		cancel()
		return nil, nil, err
	}

	// 注册服务
	err = ServiceRegister(serviceName)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	return lockLostCh, cancel, nil
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
