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

package protocol

import (
	"fmt"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
)

//go:generate mockgen -source invoker.go -destination mock/mock_invoker.go  -self_package github.com/apache/dubbo-go/protocol/mock --package mock  Invoker
// Extension - Invoker
type Invoker interface {
	common.Node
	Invoke(Invocation) Result
}

/////////////////////////////
// base invoker
/////////////////////////////

type BaseInvoker struct {
	url        common.URL
	attachment map[string]string
	available  bool
	destroyed  bool
}

var (
	attachmentKey = []string{constant.INTERFACE_KEY, constant.GROUP_KEY, constant.TOKEN_KEY, constant.TIMEOUT_KEY}
)

func NewBaseInvoker(url common.URL) *BaseInvoker {
	v := url.GetParam(constant.TOKEN_KEY, "")
	fmt.Printf("NewBaseInvoker token: %s", v)

	// server端, 在 url.SubUrl 的 baseUrl 是有的
	attachment := make(map[string]string, 0)
	for _, k := range attachmentKey {
		if v = url.GetParam(k, ""); len(v) > 0 {
			attachment[k] = v
		}
	}

	return &BaseInvoker{
		url:        url,
		attachment: attachment,
		available:  true,
		destroyed:  false,
	}
}

func (bi *BaseInvoker) GetUrl() common.URL {
	return bi.url
}

func (bi *BaseInvoker) IsAvailable() bool {
	return bi.available
}

func (bi *BaseInvoker) IsDestroyed() bool {
	return bi.destroyed
}

func (bi *BaseInvoker) Invoke(invocation Invocation) Result {
	if len(bi.attachment) > 0 {
		rpcInvocation := invocation.(*RPCInvocation)
		for k, v := range bi.attachment {
			rpcInvocation.SetAttachments(k, v)
		}
	}

	return bi.DoInvoke(invocation)
}

func (bi *BaseInvoker) DoInvoke(invocation Invocation) Result {
	return &RPCResult{}
}

func (bi *BaseInvoker) Destroy() {
	logger.Infof("Destroy invoker: %s", bi.GetUrl().String())
	bi.destroyed = true
	bi.available = false
}
