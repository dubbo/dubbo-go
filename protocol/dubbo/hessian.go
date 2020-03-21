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

package dubbo

import (
	"math"
	"strconv"
	"strings"
	"time"
)

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/apache/dubbo-go-hessian2/java_exception"
	"github.com/pkg/errors"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/extension"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/protocol/dubbo/impl"
)

type Object interface{}

type HessianSerializer struct {
}

func (h HessianSerializer) Marshal(p DubboPackage) ([]byte, error) {
	encoder := hessian.NewEncoder()
	if p.IsRequest() {
		return marshalRequest(encoder, p)
	}
	return marshalResponse(encoder, p)
}

func (h HessianSerializer) Unmarshal(input []byte, p *DubboPackage) error {
	if p.IsHeartBeat() {
		return nil
	}
	if p.IsRequest() {
		return unmarshalRequestBody(input, p)
	}
	return unmarshalResponseBody(input, p)
}

func marshalResponse(encoder *hessian.Encoder, p DubboPackage) ([]byte, error) {
	header := p.Header
	response := impl.EnsureResponsePayload(p.Body)
	if header.ResponseStatus == impl.Response_OK {
		if p.IsHeartBeat() {
			encoder.Encode(nil)
		} else {
			atta := isSupportResponseAttachment(response.Attachments[impl.DUBBO_VERSION_KEY])

			var resWithException, resValue, resNullValue int32
			if atta {
				resWithException = impl.RESPONSE_WITH_EXCEPTION_WITH_ATTACHMENTS
				resValue = impl.RESPONSE_VALUE_WITH_ATTACHMENTS
				resNullValue = impl.RESPONSE_NULL_VALUE_WITH_ATTACHMENTS
			} else {
				resWithException = impl.RESPONSE_WITH_EXCEPTION
				resValue = impl.RESPONSE_VALUE
				resNullValue = impl.RESPONSE_NULL_VALUE
			}

			if response.Exception != nil { // throw error
				encoder.Encode(resWithException)
				if t, ok := response.Exception.(java_exception.Throwabler); ok {
					encoder.Encode(t)
				} else {
					encoder.Encode(java_exception.NewThrowable(response.Exception.Error()))
				}
			} else {
				if response.RspObj == nil {
					encoder.Encode(resNullValue)
				} else {
					encoder.Encode(resValue)
					encoder.Encode(response.RspObj) // result
				}
			}

			if atta {
				encoder.Encode(response.Attachments) // attachments
			}
		}
	} else {
		if response.Exception != nil { // throw error
			encoder.Encode(response.Exception.Error())
		} else {
			encoder.Encode(response.RspObj)
		}
	}
	bs := encoder.Buffer()
	// encNull
	bs = append(bs, byte('N'))
	return bs, nil
}

func marshalRequest(encoder *hessian.Encoder, p DubboPackage) ([]byte, error) {
	service := p.Service
	request := impl.EnsureRequestPayload(p.Body)
	encoder.Encode(impl.DEFAULT_DUBBO_PROTOCOL_VERSION)
	encoder.Encode(service.Path)
	encoder.Encode(service.Version)
	encoder.Encode(service.Method)

	args, ok := request.Params.([]interface{})

	if !ok {
		logger.Infof("request args are: %+v", request.Params)
		return nil, errors.Errorf("@params is not of type: []interface{}")
	}
	types, err := hessian.GetArgsTypeList(args)
	if err != nil {
		return nil, errors.Wrapf(err, " PackRequest(args:%+v)", args)
	}
	encoder.Encode(types)
	for _, v := range args {
		encoder.Encode(v)
	}

	request.Attachments[impl.PATH_KEY] = service.Path
	request.Attachments[impl.VERSION_KEY] = service.Version
	if len(service.Group) > 0 {
		request.Attachments[impl.GROUP_KEY] = service.Group
	}
	if len(service.Interface) > 0 {
		request.Attachments[impl.INTERFACE_KEY] = service.Interface
	}
	if service.Timeout != 0 {
		request.Attachments[impl.TIMEOUT_KEY] = strconv.Itoa(int(service.Timeout / time.Millisecond))
	}

	encoder.Encode(request.Attachments)
	return encoder.Buffer(), nil

}

var versionInt = make(map[string]int)

// https://github.com/apache/dubbo/blob/dubbo-2.7.1/dubbo-common/src/main/java/org/apache/dubbo/common/Version.java#L96
// isSupportResponseAttachment is for compatibility among some dubbo version
func isSupportResponseAttachment(version string) bool {
	if version == "" {
		return false
	}

	v, ok := versionInt[version]
	if !ok {
		v = version2Int(version)
		if v == -1 {
			return false
		}
	}

	if v >= 2001000 && v <= 2060200 { // 2.0.10 ~ 2.6.2
		return false
	}
	return v >= impl.LOWEST_VERSION_FOR_RESPONSE_ATTACHMENT
}

func version2Int(version string) int {
	var v = 0
	varr := strings.Split(version, ".")
	length := len(varr)
	for key, value := range varr {
		v0, err := strconv.Atoi(value)
		if err != nil {
			return -1
		}
		v += v0 * int(math.Pow10((length-key-1)*2))
	}
	if length == 3 {
		return v * 100
	}
	return v
}

func unmarshalRequestBody(body []byte, p *DubboPackage) error {
	if p.Body == nil {
		p.SetBody(make([]interface{}, 7))
	}
	decoder := hessian.NewDecoder(body)
	var (
		err                                                     error
		dubboVersion, target, serviceVersion, method, argsTypes interface{}
		args                                                    []interface{}
	)
	req, ok := p.Body.([]interface{})
	if !ok {
		return errors.Errorf("@reqObj is not of type: []interface{}")
	}
	dubboVersion, err = decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}
	req[0] = dubboVersion

	target, err = decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}
	req[1] = target

	serviceVersion, err = decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}
	req[2] = serviceVersion

	method, err = decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}
	req[3] = method

	argsTypes, err = decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}
	req[4] = argsTypes

	ats := hessian.DescRegex.FindAllString(argsTypes.(string), -1)
	var arg interface{}
	for i := 0; i < len(ats); i++ {
		arg, err = decoder.Decode()
		if err != nil {
			return errors.WithStack(err)
		}
		args = append(args, arg)
	}
	req[5] = args

	attachments, err := decoder.Decode()
	if err != nil {
		return errors.WithStack(err)
	}

	if v, ok := attachments.(map[interface{}]interface{}); ok {
		v[impl.DUBBO_VERSION_KEY] = dubboVersion
		req[6] = hessian.ToMapStringString(v)
		buildServerSidePackageBody(p)
		return nil
	}
	return errors.Errorf("get wrong attachments: %+v", attachments)
}

func unmarshalResponseBody(body []byte, p *DubboPackage) error {
	decoder := hessian.NewDecoder(body)
	rspType, err := decoder.Decode()
	if p.Body == nil {
		p.SetBody(&impl.ResponsePayload{})
	}
	if err != nil {
		return errors.WithStack(err)
	}
	response := impl.EnsureResponsePayload(p.Body)

	switch rspType {
	case impl.RESPONSE_WITH_EXCEPTION, impl.RESPONSE_WITH_EXCEPTION_WITH_ATTACHMENTS:
		expt, err := decoder.Decode()
		if err != nil {
			return errors.WithStack(err)
		}
		if rspType == impl.RESPONSE_WITH_EXCEPTION_WITH_ATTACHMENTS {
			attachments, err := decoder.Decode()
			if err != nil {
				return errors.WithStack(err)
			}
			if v, ok := attachments.(map[interface{}]interface{}); ok {
				atta := hessian.ToMapStringString(v)
				response.Attachments = atta
			} else {
				return errors.Errorf("get wrong attachments: %+v", attachments)
			}
		}

		if e, ok := expt.(error); ok {
			response.Exception = e
		} else {
			response.Exception = errors.Errorf("got exception: %+v", expt)
		}
		return nil

	case impl.RESPONSE_VALUE, impl.RESPONSE_VALUE_WITH_ATTACHMENTS:
		rsp, err := decoder.Decode()
		if err != nil {
			return errors.WithStack(err)
		}
		if rspType == impl.RESPONSE_VALUE_WITH_ATTACHMENTS {
			attachments, err := decoder.Decode()
			if err != nil {
				return errors.WithStack(err)
			}
			if v, ok := attachments.(map[interface{}]interface{}); ok {
				atta := hessian.ToMapStringString(v)
				response.Attachments = atta
			} else {
				return errors.Errorf("get wrong attachments: %+v", attachments)
			}
		}

		return errors.WithStack(hessian.ReflectResponse(rsp, response.RspObj))

	case impl.RESPONSE_NULL_VALUE, impl.RESPONSE_NULL_VALUE_WITH_ATTACHMENTS:
		if rspType == impl.RESPONSE_NULL_VALUE_WITH_ATTACHMENTS {
			attachments, err := decoder.Decode()
			if err != nil {
				return errors.WithStack(err)
			}
			if v, ok := attachments.(map[interface{}]interface{}); ok {
				atta := hessian.ToMapStringString(v)
				response.Attachments = atta
			} else {
				return errors.Errorf("get wrong attachments: %+v", attachments)
			}
		}
		return nil
	}
	return nil
}

func buildServerSidePackageBody(pkg *DubboPackage) {
	req := pkg.GetBody().([]interface{}) // length of body should be 7
	if len(req) > 0 {
		var dubboVersion, argsTypes string
		var args []interface{}
		var attachments map[string]string
		svc := Service{}
		if req[0] != nil {
			dubboVersion = req[0].(string)
		}
		if req[1] != nil {
			svc.Path = req[1].(string)
		}
		if req[2] != nil {
			svc.Version = req[2].(string)
		}
		if req[3] != nil {
			svc.Method = req[3].(string)
		}
		if req[4] != nil {
			argsTypes = req[4].(string)
		}
		if req[5] != nil {
			args = req[5].([]interface{})
		}
		if req[6] != nil {
			attachments = req[6].(map[string]string)
		}
		if svc.Path == "" && len(attachments[constant.PATH_KEY]) > 0 {
			svc.Path = attachments[constant.PATH_KEY]
		}
		if _, ok := attachments[constant.INTERFACE_KEY]; ok {
			svc.Interface = attachments[constant.INTERFACE_KEY]
		} else {
			svc.Interface = svc.Path
		}
		if len(attachments[constant.GROUP_KEY]) > 0 {
			svc.Group = attachments[constant.GROUP_KEY]
		}
		pkg.SetService(svc)
		pkg.SetBody(map[string]interface{}{
			"dubboVersion": dubboVersion,
			"argsTypes":    argsTypes,
			"args":         args,
			"service":      common.ServiceMap.GetService(DUBBO, svc.Path), // path as a key
			"attachments":  attachments,
		})
	}
}

func init() {
	extension.SetSerializer("hessian2", HessianSerializer{})
}
