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

// Code generated by MockGen. DO NOT EDIT.
// Source: invoker.go

// Package mock is a generated GoMock package.
package mock

import (
	"context"
	"reflect"
)

import (
	"github.com/golang/mock/gomock"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/protocol"
)

// MockInvoker is a mock of Invoker interface
type MockInvoker struct {
	ctrl     *gomock.Controller
	recorder *MockInvokerMockRecorder
}

// MockInvokerMockRecorder is the mock recorder for MockInvoker
type MockInvokerMockRecorder struct {
	mock *MockInvoker
}

// NewMockInvoker creates a new mock instance
func NewMockInvoker(ctrl *gomock.Controller) *MockInvoker {
	mock := &MockInvoker{ctrl: ctrl}
	mock.recorder = &MockInvokerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInvoker) EXPECT() *MockInvokerMockRecorder {
	return m.recorder
}

// GetURL mocks base method
func (m *MockInvoker) GetURL() *common.URL {
	ret := m.ctrl.Call(m, "GetURL")
	ret0, _ := ret[0].(*common.URL)
	return ret0
}

// GetURL indicates an expected call of GetURL
func (mr *MockInvokerMockRecorder) GetURL() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockInvoker)(nil).GetURL))
}

// IsAvailable mocks base method
func (m *MockInvoker) IsAvailable() bool {
	ret := m.ctrl.Call(m, "IsAvailable")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAvailable indicates an expected call of IsAvailable
func (mr *MockInvokerMockRecorder) IsAvailable() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAvailable", reflect.TypeOf((*MockInvoker)(nil).IsAvailable))
}

// Destroy mocks base method
func (m *MockInvoker) Destroy() {
	m.ctrl.Call(m, "Destroy")
}

// Destroy indicates an expected call of Destroy
func (mr *MockInvokerMockRecorder) Destroy() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockInvoker)(nil).Destroy))
}

// Invoke mocks base method
func (m *MockInvoker) Invoke(ctx context.Context, arg0 protocol.Invocation) protocol.Result {
	ret := m.ctrl.Call(m, "Invoke", arg0)
	ret0, _ := ret[0].(protocol.Result)
	return ret0
}

// Invoke indicates an expected call of Invoke
func (mr *MockInvokerMockRecorder) Invoke(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Invoke", reflect.TypeOf((*MockInvoker)(nil).Invoke), arg0)
}
