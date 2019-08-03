package nacos

import (
	"context"
	"encoding/json"
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strconv"
	"testing"
)

func Test_Register(t *testing.T) {
	regurl, _ := common.NewURL(context.TODO(), "registry://console.nacos.io:80", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	urlMap := url.Values{}
	urlMap.Set(constant.GROUP_KEY, "guangzhou-idc")
	urlMap.Set(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER))
	urlMap.Set(constant.INTERFACE_KEY, "com.ikurento.user.UserProvider")
	urlMap.Set(constant.VERSION_KEY, "1.0.0")
	urlMap.Set(constant.CLUSTER_KEY, "mock")
	url, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParams(urlMap), common.WithMethods([]string{"GetUser", "AddUser"}))

	reg, err := newNacosRegistry(&regurl)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("new nacos registry error:%s \n", err.Error())
		return
	}
	err = reg.Register(url)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("register error:%s \n", err.Error())
		return
	}
	nacosReg := reg.(*nacosRegistry)
	service, _ := nacosReg.namingClient.GetService(vo.GetServiceParam{ServiceName: "providers:com.ikurento.user.UserProvider:1.0.0:guangzhou-idc"})
	data, _ := json.Marshal(service)
	logger.Infof(string(data))
	assert.Equal(t, 1, len(service.Hosts))
}

func TestNacosRegistry_Subscribe(t *testing.T) {
	regurl, _ := common.NewURL(context.TODO(), "registry://console.nacos.io:80", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	urlMap := url.Values{}
	urlMap.Set(constant.GROUP_KEY, "guangzhou-idc")
	urlMap.Set(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER))
	urlMap.Set(constant.INTERFACE_KEY, "com.ikurento.user.UserProvider")
	urlMap.Set(constant.VERSION_KEY, "1.0.0")
	urlMap.Set(constant.CLUSTER_KEY, "mock")
	urlMap.Set(constant.NACOS_PATH_KEY, "")
	url, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParams(urlMap), common.WithMethods([]string{"GetUser", "AddUser"}))

	reg, _ := newNacosRegistry(&regurl)
	err := reg.Register(url)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("new nacos registry error:%s \n", err.Error())
		return
	}

	regurl.Params.Set(constant.ROLE_KEY, strconv.Itoa(common.CONSUMER))
	reg2, _ := newNacosRegistry(&regurl)
	listener, err := reg2.Subscribe(url)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("subscribe error:%s \n", err.Error())
		return
	}
	serviceEvent, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		logger.Errorf("listener error:%s \n", err.Error())
		return
	}
	logger.Infof("serviceEvent:%+v \n", serviceEvent)
	assert.Regexp(t, ".*ServiceEvent{Action{add}.*", serviceEvent.String())

}

func TestNacosRegistry_Subscribe_del(t *testing.T) {
	regurl, _ := common.NewURL(context.TODO(), "registry://console.nacos.io:80", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	urlMap := url.Values{}
	urlMap.Set(constant.GROUP_KEY, "guangzhou-idc")
	urlMap.Set(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER))
	urlMap.Set(constant.INTERFACE_KEY, "com.ikurento.user.UserProvider")
	urlMap.Set(constant.VERSION_KEY, "1.0.0")
	urlMap.Set(constant.CLUSTER_KEY, "mock")
	urlMap.Set(constant.NACOS_PATH_KEY, "")
	url1, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider", common.WithParams(urlMap), common.WithMethods([]string{"GetUser", "AddUser"}))
	url2, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.2:20000/com.ikurento.user.UserProvider", common.WithParams(urlMap), common.WithMethods([]string{"GetUser", "AddUser"}))

	reg, _ := newNacosRegistry(&regurl)
	err := reg.Register(url1)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("register1 error:%s \n", err.Error())
		return
	}
	err = reg.Register(url2)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("register2 error:%s \n", err.Error())
		return
	}

	regurl.Params.Set(constant.ROLE_KEY, strconv.Itoa(common.CONSUMER))
	reg2, _ := newNacosRegistry(&regurl)
	listener, err := reg2.Subscribe(url1)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("subscribe error:%s \n", err.Error())
		return
	}

	serviceEvent1, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		logger.Errorf("listener1 error:%s \n", err.Error())
		return
	}
	logger.Infof("serviceEvent1:%+v \n", serviceEvent1)
	assert.Regexp(t, ".*ServiceEvent{Action{add}.*", serviceEvent1.String())

	serviceEvent2, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		logger.Errorf("listener2 error:%s \n", err.Error())
		return
	}
	logger.Infof("serviceEvent2:%+v \n", serviceEvent2)
	assert.Regexp(t, ".*ServiceEvent{Action{add}.*", serviceEvent2.String())

	nacosReg := reg.(*nacosRegistry)
	//手动注销实例
	nacosReg.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{Ip: "127.0.0.2", Port: 20000, ServiceName: "providers:com.ikurento.user.UserProvider:1.0.0:guangzhou-idc"})

	serviceEvent3, _ := listener.Next()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	logger.Infof("serviceEvent3:%+v \n", serviceEvent3)
	assert.Regexp(t, ".*ServiceEvent{Action{delete}.*", serviceEvent3.String())
}

func TestNacosListener_Close(t *testing.T) {
	regurl, _ := common.NewURL(context.TODO(), "registry://console.nacos.io:80", common.WithParamsValue(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER)))
	urlMap := url.Values{}
	urlMap.Set(constant.GROUP_KEY, "guangzhou-idc")
	urlMap.Set(constant.ROLE_KEY, strconv.Itoa(common.PROVIDER))
	urlMap.Set(constant.INTERFACE_KEY, "com.ikurento.user.UserProvider2")
	urlMap.Set(constant.VERSION_KEY, "1.0.0")
	urlMap.Set(constant.CLUSTER_KEY, "mock")
	urlMap.Set(constant.NACOS_PATH_KEY, "")
	url1, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000/com.ikurento.user.UserProvider2", common.WithParams(urlMap), common.WithMethods([]string{"GetUser", "AddUser"}))
	reg, _ := newNacosRegistry(&regurl)
	listener, err := reg.Subscribe(url1)
	assert.Nil(t, err)
	if err != nil {
		logger.Errorf("subscribe error:%s \n", err.Error())
		return
	}
	listener.Close()
	_, err = listener.Next()
	assert.NotNil(t, err)
}
