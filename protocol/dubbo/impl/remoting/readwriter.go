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

// TODO: zero.xu readwrite 中将client/server handler 分开
package remoting

import (
	"bufio"
	"bytes"
	"reflect"
)

import (
	hessian "github.com/apache/dubbo-go-hessian2"
	"github.com/dubbogo/getty"
	perrors "github.com/pkg/errors"
)

import (
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/protocol/dubbo/impl"
)

////////////////////////////////////////////
// RpcServerPackageHandler
////////////////////////////////////////////

var (
	rpcServerPkgHandler = &RpcServerPackageHandler{}
)

// RpcServerPackageHandler ...
type RpcServerPackageHandler struct {
}

func (p *RpcServerPackageHandler) Read(ss getty.Session, data []byte) (interface{}, int, error) {
	pkg := NewServerRequestPackage(data)
	if err := pkg.ReadHeader(); err != nil {
		originErr := perrors.Cause(err)
		if originErr == hessian.ErrHeaderNotEnough || originErr == hessian.ErrBodyNotEnough {
			return nil, 0, nil
		}
		return nil, 0, perrors.WithStack(err)
	}

	if pkg.IsHeartBeat() {
		return pkg, pkg.GetLen(), nil
	}

	if err := impl.LoadSerializer(pkg); err != nil {
		return nil, 0, err
	}

	err := pkg.Unmarshal()
	if err != nil {
		originErr := perrors.Cause(err)
		if originErr == hessian.ErrHeaderNotEnough || originErr == hessian.ErrBodyNotEnough {
			return nil, 0, nil
		}

		logger.Errorf("pkg.Unmarshal(ss:%+v, len(@data):%d) = error:%+v", ss, len(data), err)
		return nil, 0, perrors.WithStack(err)
	}
	return pkg, pkg.GetLen(), nil
}

func (p *RpcServerPackageHandler) Write(ss getty.Session, pkg interface{}) ([]byte, error) {
	res, ok := pkg.(*impl.DubboPackage)
	if !ok {
		logger.Errorf("illegal pkg:%+v\n, it is %+v", pkg, reflect.TypeOf(pkg))
		return nil, perrors.New("invalid rpc response")
	}
	buf, err := res.Marshal()
	if err != nil {
		logger.Warnf("binary.Write(res{%#v}) = err{%#v}", res, perrors.WithStack(err))
		return nil, perrors.WithStack(err)
	}
	return buf.Bytes(), nil
}

// server side receive request package, just for deserialization
func NewServerRequestPackage(data []byte) *impl.DubboPackage {
	return &impl.DubboPackage{
		Header:  impl.DubboHeader{},
		Service: impl.Service{},
		Body:    make([]interface{}, 7),
		Err:     nil,
		Codec:   impl.NewDubboCodec(bufio.NewReaderSize(bytes.NewBuffer(data), len(data))),
	}

}
