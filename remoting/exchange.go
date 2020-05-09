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
package remoting

import (
	"time"

	"github.com/apache/dubbo-go/common"
	"go.uber.org/atomic"
)

var (
	sequence atomic.Uint64
)

func init() {
	sequence.Store(0)
}

func SequenceId() uint64 {
	return sequence.Add(2)
}

// Request ...
type Request struct {
	Id       int64
	Version  string
	SerialID byte
	Data     interface{}
	TwoWay   bool
	Event    bool
	broken   bool
}

// NewRequest ...
func NewRequest(version string) *Request {
	return &Request{
		Id:      int64(SequenceId()),
		Version: version,
	}
}

// Response ...
type Response struct {
	Id       int64
	Version  string
	SerialID byte
	Status   uint8
	Event    bool
	Error    error
	Result   interface{}
}

// NewResponse ...
func NewResponse(id int64, version string) *Response {
	return &Response{
		Id:      id,
		Version: version,
	}
}

func (response *Response) IsHeartbeat() bool {
	return response.Event && response.Result == nil
}

type Options struct {
	// connect timeout
	ConnectTimeout time.Duration
}

//AsyncCallbackResponse async response for dubbo
type AsyncCallbackResponse struct {
	common.CallbackResponse
	Opts      Options
	Cause     error
	Start     time.Time // invoke(call) start time == write start time
	ReadStart time.Time // read start time, write duration = ReadStart - Start
	Reply     interface{}
}

type PendingResponse struct {
	seq       int64
	Err       error
	start     time.Time
	ReadStart time.Time
	Callback  common.AsyncCallback
	response  *Response
	Reply     interface{}
	Done      chan struct{}
}

// NewPendingResponse ...
func NewPendingResponse(id int64) *PendingResponse {
	return &PendingResponse{
		seq:      id,
		start:    time.Now(),
		response: &Response{},
		Done:     make(chan struct{}),
	}
}

func (r *PendingResponse) SetResponse(response *Response) {
	r.response = response
}

// GetCallResponse ...
func (r PendingResponse) GetCallResponse() common.CallbackResponse {
	return AsyncCallbackResponse{
		Cause:     r.Err,
		Start:     r.start,
		ReadStart: r.ReadStart,
		Reply:     r.response,
	}
}
