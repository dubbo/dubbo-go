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

package consul

import (
	"github.com/apache/dubbo-go/common"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/extension"
	"github.com/apache/dubbo-go/common/observer"
	"github.com/apache/dubbo-go/common/observer/dispatcher"
	"github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/registry"
)

var (
	testName    = "test"
	registryURL = common.URL{
		Path:     "",
		Username: "",
		Password: "",
		Methods:  nil,
		SubURL:   nil,
	}
)

func TestConsulServiceDiscovery_newConsulServiceDiscovery(t *testing.T) {
	name := "consul1"
	_, err := newConsulServiceDiscovery(name)
	assert.NotNil(t, err)

	sdc := &config.ServiceDiscoveryConfig{
		Protocol:  "consul",
		RemoteRef: "mock",
	}

	config.GetBaseConfig().ServiceDiscoveries[name] = sdc

	_, err = newConsulServiceDiscovery(name)
	assert.NotNil(t, err)

	config.GetBaseConfig().Remotes["mock"] = &config.RemoteConfig{
		Address: "", // TODO
	}

	res, err := newConsulServiceDiscovery(name)
	assert.Nil(t, err)
	assert.NotNil(t, res)
}

func TestConsulServiceDiscovery_Destroy(t *testing.T) {
	prepareData()
	serviceDiscovery, err := extension.GetServiceDiscovery(constant.CONSUL_KEY, testName)
	_, registryUrl := prepareService()
	serviceDiscovery.Initialize(registryUrl)
	assert.Nil(t, err)
	assert.NotNil(t, serviceDiscovery)
	err = serviceDiscovery.Destroy()
	assert.Nil(t, err)
	assert.Nil(t, serviceDiscovery.(*consulServiceDiscovery).consulClient)
}

func TestConsulServiceDiscovery_CRUD(t *testing.T) {
	prepareData()
	extension.SetEventDispatcher("mock", func() observer.EventDispatcher {
		return &dispatcher.MockEventDispatcher{}
	})

	extension.SetAndInitGlobalDispatcher("mock")
	rand.Seed(time.Now().Unix())

	instance, registryUrl := prepareService()

	// clean data
	serviceDiscovery, err := extension.GetServiceDiscovery(constant.CONSUL_KEY, testName)
	assert.Nil(t, err)

	err = serviceDiscovery.Initialize(registryUrl)
	assert.Nil(t, err)
	// clean data for local test
	err = serviceDiscovery.Unregister(instance)
	assert.Nil(t, err)

	err = serviceDiscovery.Register(instance)
	assert.Nil(t, err)

	//sometimes nacos may be failed to push update of instance,
	//so it need 10s to pull, we sleep 10 second to make sure instance has been update
	time.Sleep(11 * time.Second)
	page := serviceDiscovery.GetHealthyInstancesByPage(instance.GetServiceName(), 0, 10, true)
	assert.NotNil(t, page)
	assert.Equal(t, 0, page.GetOffset())
	assert.Equal(t, 10, page.GetPageSize())
	assert.Equal(t, 1, page.GetDataSize())

	instance = page.GetData()[0].(*registry.DefaultServiceInstance)
	assert.NotNil(t, instance)
	assert.Equal(t, buildID(instance), instance.GetId())
	assert.Equal(t, instance.GetHost(), instance.GetHost())
	assert.Equal(t, instance.GetPort(), instance.GetPort())
	assert.Equal(t, instance.GetServiceName(), instance.GetServiceName())
	assert.Equal(t, 0, len(instance.GetMetadata()))

	instance.GetMetadata()["a"] = "b"
	err = serviceDiscovery.Update(instance)
	assert.Nil(t, err)

	time.Sleep(11 * time.Second)
	pageMap := serviceDiscovery.GetRequestInstances([]string{instance.GetServiceName()}, 0, 1)
	assert.Equal(t, 1, len(pageMap))

	page = pageMap[instance.GetServiceName()]
	assert.NotNil(t, page)
	assert.Equal(t, 1, len(page.GetData()))

	instance = page.GetData()[0].(*registry.DefaultServiceInstance)
	v, _ := instance.GetMetadata()["a"]
	assert.Equal(t, "b", v)

	// test dispatcher event
	err = serviceDiscovery.DispatchEventByServiceName(instance.GetServiceName())
	assert.Nil(t, err)

	// test AddListener
	err = serviceDiscovery.AddListener(&registry.ServiceInstancesChangedListener{})
	assert.Nil(t, err)
}

func prepareData() {
	config.GetBaseConfig().ServiceDiscoveries[testName] = &config.ServiceDiscoveryConfig{
		Protocol:  "consul",
		RemoteRef: testName,
	}

	config.GetBaseConfig().Remotes[testName] = &config.RemoteConfig{
		Address:    "", // TODO
		TimeoutStr: "10s",
	}
}
func prepareService() (registry.ServiceInstance, common.URL) {
	serviceName := "service-name" + strconv.Itoa(rand.Intn(10000))
	id := "id"
	host := "host"
	port := 123

	registryUrl, _ := common.NewURL("dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider?anyhost=true&" +
		"application=BDTService&category=providers&default.timeout=10000&dubbo=dubbo-provider-golang-1.0.0&" +
		"environment=dev&interface=com.ikurento.user.UserProvider&ip=192.168.56.1&methods=GetUser%2C&" +
		"module=dubbogo+user-info+server&org=ikurento.com&owner=ZX&pid=1447&revision=0.0.1&" +
		"side=provider&timeout=3000&timestamp=1556509797245&consul-check-pass-interval=17000&consul-deregister-critical-service-after=20s&" +
		"consul-watch-timeout=60000")

	return &registry.DefaultServiceInstance{
		Id:          id,
		ServiceName: serviceName,
		Host:        host,
		Port:        port,
		Enable:      true,
		Healthy:     true,
		Metadata:    nil,
	}, registryUrl
}
