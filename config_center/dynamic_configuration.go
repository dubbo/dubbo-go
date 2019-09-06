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

package config_center

import (
	"time"
)
import (
	"github.com/apache/dubbo-go/remoting"
)

//////////////////////////////////////////
// DynamicConfiguration
//////////////////////////////////////////
const DEFAULT_GROUP = "dubbo"
const DEFAULT_CONFIG_TIMEOUT = "10s"

type DynamicConfiguration interface {
	Parser() ConfigurationParser
	SetParser(ConfigurationParser)
	AddListener(string, remoting.ConfigurationListener, ...Option)
	RemoveListener(string, remoting.ConfigurationListener, ...Option)
	GetConfig(string, ...Option) (string, error)
	SetConfig(string, string, string) error
	GetConfigs(string, ...Option) (string, error)
}

type Options struct {
	Group   string
	Timeout time.Duration
}

type Option func(*Options)

func WithGroup(group string) Option {
	return func(opt *Options) {
		opt.Group = group
	}
}

func WithTimeout(time time.Duration) Option {
	return func(opt *Options) {
		opt.Timeout = time
	}
}
