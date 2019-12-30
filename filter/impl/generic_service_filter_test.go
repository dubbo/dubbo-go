package impl

import (
	"context"
	"errors"
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/proxy/proxy_factory"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/invocation"
)

type TestStruct struct {
	AaAa string
	BaBa string `m:"baBa"`
	XxYy struct {
		xxXx string `m:"xxXx"`
		Xx   string `m:"xx"`
	} `m:"xxYy"`
}

func (c *TestStruct) JavaClassName() string {
	return "com.test.testStruct"
}

type TestService struct {
}

func (ts *TestService) MethodOne(ctx context.Context, test1 *TestStruct, test2 []TestStruct,
	test3 interface{}, test4 []interface{}, test5 *string) (*TestStruct, error) {
	if test1 == nil {
		return nil, errors.New("param test1 is nil")
	}
	if test2 == nil {
		return nil, errors.New("param test2 is nil")
	}
	if test3 == nil {
		return nil, errors.New("param test3 is nil")
	}
	if test4 == nil {
		return nil, errors.New("param test4 is nil")
	}
	if test5 == nil {
		return nil, errors.New("param test5 is nil")
	}
	return &TestStruct{}, nil
}

func (s *TestService) Reference() string {
	return "com.test.Path"
}

func TestGenericServiceFilter_Invoke(t *testing.T) {
	hessian.RegisterPOJO(&TestStruct{})
	methodName := "$invoke"
	m := make(map[string]interface{})
	m["AaAa"] = "nihao"
	x := make(map[string]interface{})
	x["xxXX"] = "nihaoxxx"
	m["XxYy"] = x
	aurguments := []interface{}{
		"MethodOne",
		nil,
		[]hessian.Object{
			hessian.Object(m),
			hessian.Object(append(make([]map[string]interface{}, 1), m)),
			hessian.Object("111"),
			hessian.Object(append(make([]map[string]interface{}, 1), m)),
			hessian.Object("222")},
	}
	s := &TestService{}
	_, _ = common.ServiceMap.Register("testprotocol", s)
	rpcInvocation := invocation.NewRPCInvocation(methodName, aurguments, nil)
	filter := GetGenericServiceFilter()
	url, _ := common.NewURL(context.Background(), "testprotocol://127.0.0.1:20000/com.test.Path")
	result := filter.Invoke(&proxy_factory.ProxyInvoker{BaseInvoker: *protocol.NewBaseInvoker(url)}, rpcInvocation)
	assert.NotNil(t, result)
	assert.Nil(t, result.Error())
}

func TestGenericServiceFilter_ResponseTestStruct(t *testing.T) {
	ts := &TestStruct{
		AaAa: "aaa",
		BaBa: "bbb",
		XxYy: struct {
			xxXx string `m:"xxXx"`
			Xx   string `m:"xx"`
		}{},
	}
	result := &protocol.RPCResult{
		Rest: ts,
	}
	aurguments := []interface{}{
		"MethodOne",
		nil,
		[]hessian.Object{nil},
	}
	filter := GetGenericServiceFilter()
	methodName := "$invoke"
	rpcInvocation := invocation.NewRPCInvocation(methodName, aurguments, nil)
	r := filter.OnResponse(result, nil, rpcInvocation)
	assert.NotNil(t, r.Result())
	assert.Equal(t, reflect.ValueOf(r.Result()).Kind(), reflect.Map)
}

func TestGenericServiceFilter_ResponseString(t *testing.T) {
	str := "111"
	result := &protocol.RPCResult{
		Rest: str,
	}
	aurguments := []interface{}{
		"MethodOne",
		nil,
		[]hessian.Object{nil},
	}
	filter := GetGenericServiceFilter()
	methodName := "$invoke"
	rpcInvocation := invocation.NewRPCInvocation(methodName, aurguments, nil)
	r := filter.OnResponse(result, nil, rpcInvocation)
	assert.NotNil(t, r.Result())
	assert.Equal(t, reflect.ValueOf(r.Result()).Kind(), reflect.String)
}
