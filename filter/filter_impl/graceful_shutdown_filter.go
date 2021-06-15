/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package filter_impl

import (
	"context"
	"sync"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common/constant"
	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/common/logger"
	"dubbo.apache.org/dubbo-go/v3/config"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol"
)

func init() {
	// `init()` is performed before config.Load(), so shutdownConfig will be retrieved after config was loaded.
	extension.SetFilter(constant.CONSUMER_SHUTDOWN_FILTER, func() filter.Filter {
		return &gracefulShutdownFilter{}
	})

	extension.SetFilter(constant.PROVIDER_SHUTDOWN_FILTER, func() filter.Filter {
		return &gracefulShutdownFilter{}
	})
}

type gracefulShutdownFilter struct {
	mutex          sync.Mutex // protect the following variables
	activeCount    int32
	shutdownConfig *config.ShutdownConfig
}

// Invoke adds the requests count and block the new requests if application is closing
func (gf *gracefulShutdownFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	if gf.rejectNewRequest() {
		logger.Info("The application is closing, new request will be rejected.")
		return gf.getRejectHandler().RejectedExecution(invoker.GetURL(), invocation)
	}
	gf.mutex.Lock()
	gf.activeCount += 1
	if gf.shutdownConfig != nil && gf.activeCount > 0 {
		gf.shutdownConfig.RequestsFinished = false
	}
	gf.mutex.Unlock()
	return invoker.Invoke(ctx, invocation)
}

// OnResponse reduces the number of active processes then return the process result
func (gf *gracefulShutdownFilter) OnResponse(ctx context.Context, result protocol.Result, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	gf.mutex.Lock()
	gf.activeCount -= 1
	if gf.shutdownConfig != nil && gf.activeCount <= 0 {
		gf.shutdownConfig.RequestsFinished = true
	}
	gf.mutex.Unlock()
	return result
}

func (gf *gracefulShutdownFilter) Set(name string, conf interface{}) {
	switch name {
	case config.GracefulShutdownFilterShutdownConfig:
		if shutdownConfig, ok := conf.(*config.ShutdownConfig); !ok {
			gf.shutdownConfig = shutdownConfig
			return
		}
		logger.Warnf("the type of config for {%s} should be *config.ShutdownConfig", config.GracefulShutdownFilterShutdownConfig)
	default:
		// do nothing
	}
}

func (gf *gracefulShutdownFilter) rejectNewRequest() bool {
	if gf.shutdownConfig == nil {
		return false
	}
	return gf.shutdownConfig.RejectRequest
}

func (gf *gracefulShutdownFilter) getRejectHandler() filter.RejectedExecutionHandler {
	handler := constant.DEFAULT_KEY
	if gf.shutdownConfig != nil && len(gf.shutdownConfig.RejectRequestHandler) > 0 {
		handler = gf.shutdownConfig.RejectRequestHandler
	}
	return extension.GetRejectedExecutionHandler(handler)
}
