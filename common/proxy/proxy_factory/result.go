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

package proxy_factory

import (
	"context"
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/common/proxy"
	"github.com/apache/dubbo-go/protocol"
	"github.com/apache/dubbo-go/protocol/invocation"
	"reflect"
)

var (
	typeOfInterface = reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem()).Type()
	typeOfResult    = reflect.Zero(reflect.TypeOf((*protocol.Result)(nil)).Elem()).Type()
)

// ResultProxyFactory proxy factory, return a protocol.Result with attachments
type ResultProxyFactory struct {
}

// NewResultProxyFactory returns a proxy factory instance
func NewResultProxyFactory(_ ...proxy.Option) proxy.ProxyFactory {
	return &ResultProxyFactory{}
}

func (f *ResultProxyFactory) GetProxy(invoker protocol.Invoker, url *common.URL) *proxy.Proxy {
	return f.GetAsyncProxy(invoker, nil, url)
}

func (f *ResultProxyFactory) GetAsyncProxy(invoker protocol.Invoker, callBack interface{}, url *common.URL) *proxy.Proxy {
	attrs := map[string]string{}
	attrs[constant.ASYNC_KEY] = url.GetParam(constant.ASYNC_KEY, "false")
	return proxy.NewProxyWithOptions(invoker, callBack, attrs,
		proxy.WithProxyImplementFunc(NewResultProxyImplFunc(attrs)))
}

// GetInvoker gets a invoker
func (f *ResultProxyFactory) GetInvoker(url *common.URL) protocol.Invoker {
	return &ProxyInvoker{
		BaseInvoker: *protocol.NewBaseInvoker(url),
	}
}

func NewResultProxyImplFunc(attr map[string]string) proxy.ImplementFunc {
	return func(p *proxy.Proxy, rpc common.RPCService) {
		serviceValue := reflect.ValueOf(rpc)
		serviceElem := serviceValue.Elem()
		serviceType := serviceElem.Type()
		numField := serviceElem.NumField()
		for i := 0; i < numField; i++ {
			f := serviceElem.Field(i)
			if !(f.Kind() == reflect.Func && f.IsValid() && f.CanSet()) {
				continue
			}
			funcField := serviceType.Field(i)
			funcName := funcField.Tag.Get("dubbo")
			if funcName == "" {
				funcName = funcField.Name
			}
			// Only generic method: Invoke/$invoke
			if funcName != "Invoke" && funcName != "$invoke" {
				continue
			}
			// Enforce param types：
			// Invoke(Context, []{Method, []{Arguments}}) protocol.Result
			inNum := funcField.Type.NumIn()
			if inNum != 2 {
				logger.Errorf("Generic func requires 2 in-arg type, func: %s(%s), was: %d",
					funcField.Name, funcField.Type.String(), inNum)
				continue
			}
			outNum := funcField.Type.NumOut()
			if outNum != 1 || funcField.Type.Out(0) != typeOfResult {
				logger.Errorf("Generic func requires 1 out-result type, func: %s(%s), was: %d",
					funcField.Name, funcField.Type.String(), outNum)
				continue
			}
			f.Set(reflect.MakeFunc(f.Type(), makeResultProxyFunc(funcName, p, attr)))
		}
	}
}

// for: func Invoke(goctx, []interface{}{service.Method, types, values}) protocol.Result;
func makeResultProxyFunc(funcName string, proxy *proxy.Proxy, usrAttr map[string]string) func([]reflect.Value) []reflect.Value {
	return func(funArgs []reflect.Value) []reflect.Value {
		// Context
		invCtx := funArgs[0].Interface().(context.Context)
		invReply := reflect.New(typeOfInterface)
		invArgs := funArgs[1].Interface().([]interface{})
		invValues := funArgs[1]
		inv := invocation.NewRPCInvocationWithOptions(
			invocation.WithMethodName(funcName),
			invocation.WithCallBack(proxy.GetCallback()),
			invocation.WithArguments(invArgs),
			invocation.WithParameterValues([]reflect.Value{invValues}),
			invocation.WithReply(invReply.Interface()),
		)
		if invCtx == nil {
			invCtx = context.Background()
		}
		// Build-in attachments
		for k, value := range usrAttr {
			inv.SetAttachments(k, value)
		}
		// User context attachments
		attachments := invCtx.Value(constant.AttachmentKey)
		if ssmap, ok := attachments.(map[string]string); ok {
			for k, v := range ssmap {
				inv.SetAttachments(k, v)
			}
		} else if simap, ok := attachments.(map[string]interface{}); ok {
			for k, v := range simap {
				inv.SetAttachments(k, v)
			}
		} else {
			logger.Errorf("Attachments requires map[string]string OR map[string]interface{}, was: %T", attachments)
		}
		// Invoke and unwrap reply value
		result := proxy.GetInvoker().Invoke(invCtx, inv)
		if nil == result.Error() {
			result.SetResult(invReply.Elem().Interface())
		}
		return []reflect.Value{reflect.ValueOf(&result).Elem()}
	}
}
