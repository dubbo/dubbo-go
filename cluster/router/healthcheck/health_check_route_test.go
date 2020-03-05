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

package healthcheck

import (
	"math"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/invocation"
)

func TestHealthCheckRouter_Route(t *testing.T) {
	defer protocol.CleanAllStatus()
	consumerURL, _ := common.NewURL("dubbo://192.168.10.1/com.ikurento.user.UserProvider")
	consumerURL.SetParam(HEALTH_ROUTE_ENABLED_KEY, "true")
	url1, _ := common.NewURL("dubbo://192.168.10.10:20000/com.ikurento.user.UserProvider")
	url2, _ := common.NewURL("dubbo://192.168.10.11:20000/com.ikurento.user.UserProvider")
	url3, _ := common.NewURL("dubbo://192.168.10.12:20000/com.ikurento.user.UserProvider")
	hcr, _ := NewHealthCheckRouter(&consumerURL)

	var invokers []protocol.Invoker
	invoker1 := NewMockInvoker(url1, 1)
	invoker2 := NewMockInvoker(url2, 1)
	invoker3 := NewMockInvoker(url3, 1)
	invokers = append(invokers, invoker1, invoker2, invoker3)
	inv := invocation.NewRPCInvocation("test", nil, nil)
	res := hcr.Route(invokers, &consumerURL, inv)
	// now all invokers are healthy
	assert.True(t, len(res) == len(invokers))

	for i := 0; i < 10; i++ {
		request(url1, "test", 0, false, false)
	}
	res = hcr.Route(invokers, &consumerURL, inv)
	// invokers1  is unhealthy now
	assert.True(t, len(res) == 2 && !contains(res, invoker1))

	for i := 0; i < 10; i++ {
		request(url1, "test", 0, false, false)
		request(url2, "test", 0, false, false)
	}

	res = hcr.Route(invokers, &consumerURL, inv)
	// only invokers3  is healthy now
	assert.True(t, len(res) == 1 && !contains(res, invoker1) && !contains(res, invoker2))

	for i := 0; i < 10; i++ {
		request(url1, "test", 0, false, false)
		request(url2, "test", 0, false, false)
		request(url3, "test", 0, false, false)
	}

	res = hcr.Route(invokers, &consumerURL, inv)
	// now all invokers are unhealthy, so downgraded to all
	assert.True(t, len(res) == 3)

	// reset the invoker1 successive failed count, so invoker1 go to healthy
	request(url1, "test", 0, false, true)
	res = hcr.Route(invokers, &consumerURL, inv)
	assert.True(t, contains(res, invoker1))

	for i := 0; i < 6; i++ {
		request(url1, "test", 0, false, false)
	}
	// now all invokers are unhealthy, so downgraded to all again
	res = hcr.Route(invokers, &consumerURL, inv)
	assert.True(t, len(res) == 3)
	time.Sleep(time.Second * 2)
	// invoker1 go to healthy again after 2s
	res = hcr.Route(invokers, &consumerURL, inv)
	assert.True(t, contains(res, invoker1))

}

func contains(invokers []protocol.Invoker, invoker protocol.Invoker) bool {
	for _, e := range invokers {
		if e == invoker {
			return true
		}
	}
	return false
}

func TestNewHealthCheckRouter(t *testing.T) {
	defer protocol.CleanAllStatus()
	url, _ := common.NewURL("dubbo://192.168.10.10:20000/com.ikurento.user.UserProvider")
	hcr, _ := NewHealthCheckRouter(&url)
	h := hcr.(*HealthCheckRouter)
	assert.Nil(t, h.checker)

	url.SetParam(HEALTH_ROUTE_ENABLED_KEY, "true")
	hcr, _ = NewHealthCheckRouter(&url)
	h = hcr.(*HealthCheckRouter)
	assert.NotNil(t, h.checker)

	dhc := h.checker.(*DefaultHealthChecker)
	assert.Equal(t, dhc.outStandingRequestConutLimit, int32(math.MaxInt32))
	assert.Equal(t, dhc.requestSuccessiveFailureThreshold, int32(DEFAULT_SUCCESSIVE_FAILED_THRESHOLD))
	assert.Equal(t, dhc.circuitTrippedTimeoutFactor, int32(DEFAULT_CIRCUIT_TRIPPED_TIMEOUT_FACTOR))

	url.SetParam(CIRCUIT_TRIPPED_TIMEOUT_FACTOR_KEY, "500")
	url.SetParam(SUCCESSIVE_FAILED_REQUEST_THRESHOLD_KEY, "10")
	url.SetParam(OUTSTANDING_REQUEST_COUNT_LIMIT_KEY, "1000")
	hcr, _ = NewHealthCheckRouter(&url)
	h = hcr.(*HealthCheckRouter)
	dhc = h.checker.(*DefaultHealthChecker)
	assert.Equal(t, dhc.outStandingRequestConutLimit, int32(1000))
	assert.Equal(t, dhc.requestSuccessiveFailureThreshold, int32(10))
	assert.Equal(t, dhc.circuitTrippedTimeoutFactor, int32(500))
}
