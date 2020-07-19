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
// Source: tps_limit_strategy.go

// Package filter is a generated GoMock package.
package tps

import (
	gomock "github.com/golang/mock/gomock"
)

import (
	"reflect"
)

// MockTpsLimitStrategy is a mock of TpsLimitStrategy interface
type MockTpsLimitStrategy struct {
	ctrl     *gomock.Controller
	recorder *MockTpsLimitStrategyMockRecorder
}

// MockTpsLimitStrategyMockRecorder is the mock recorder for MockTpsLimitStrategy
type MockTpsLimitStrategyMockRecorder struct {
	mock *MockTpsLimitStrategy
}

// NewMockTpsLimitStrategy creates a new mock instance
func NewMockTpsLimitStrategy(ctrl *gomock.Controller) *MockTpsLimitStrategy {
	mock := &MockTpsLimitStrategy{ctrl: ctrl}
	mock.recorder = &MockTpsLimitStrategyMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTpsLimitStrategy) EXPECT() *MockTpsLimitStrategyMockRecorder {
	return m.recorder
}

// IsAllowable mocks base method
func (m *MockTpsLimitStrategy) IsAllowable() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAllowable")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsAllowable indicates an expected call of IsAllowable
func (mr *MockTpsLimitStrategyMockRecorder) IsAllowable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAllowable", reflect.TypeOf((*MockTpsLimitStrategy)(nil).IsAllowable))
}
