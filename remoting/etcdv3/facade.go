package etcdv3

import (
	"sync"
	"time"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
)

import (
	"github.com/dubbogo/getty"
	perrors "github.com/pkg/errors"
)

type clientFacade interface {
	Client() *Client
	SetClient(*Client)
	ClientLock() *sync.Mutex
	WaitGroup() *sync.WaitGroup //for wait group control, zk client listener & zk client container
	GetDone() chan struct{}     //for zk client control
	RestartCallBack() bool
	common.Node
}

func HandleClientRestart(r clientFacade) {

	var (
		err       error
		failTimes int
	)

	defer r.WaitGroup().Done()
LOOP:
	for {
		select {
		case <-r.GetDone():
			logger.Warnf("(ETCDV3ProviderRegistry)reconnectETCDV3 goroutine exit now...")
			break LOOP
			// re-register all services
		case <-r.Client().Done():
			r.ClientLock().Lock()
			clientName := RegistryETCDV3Client
			timeout, _ := time.ParseDuration(r.GetUrl().GetParam(constant.REGISTRY_TIMEOUT_KEY, constant.DEFAULT_REG_TIMEOUT))
			endpoint := r.GetUrl().Location
			r.Client().Close()
			r.SetClient(nil)
			r.ClientLock().Unlock()

			// 接zk，直至成功
			failTimes = 0
			for {
				select {
				case <-r.GetDone():
					logger.Warnf("(ETCDV3ProviderRegistry)reconnectETCDRegistry goroutine exit now...")
					break LOOP
				case <-getty.GetTimeWheel().After(timeSecondDuration(failTimes * ConnDelay)): // 防止疯狂重连etcd
				}
				err = ValidateClient(r,
					WithName(clientName),
					WithEndpoints(endpoint),
					WithTimeout(timeout),
				)
				logger.Infof("ETCDV3ProviderRegistry.validateETCDV3Client(etcd Addr{%s}) = error{%#v}",
					endpoint, perrors.WithStack(err))
				if err == nil {
					if r.RestartCallBack() {
						break
					}
				}
				failTimes++
				if MaxFailTimes <= failTimes {
					failTimes = MaxFailTimes
				}
			}
		}
	}
}
