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

package provider

import (
	"dubbo.apache.org/dubbo-go/v3/config/protocol"
	"dubbo.apache.org/dubbo-go/v3/config/registry"
	"dubbo.apache.org/dubbo-go/v3/config/service"
	"dubbo.apache.org/dubbo-go/v3/config/shutdown"
)

import (
	"github.com/creasty/defaults"
)

import (
	"dubbo.apache.org/dubbo-go/v3/common/constant"
)

// ProviderConfig is the default configuration of service provider
type ProviderConfig struct {
	//base.Config         `yaml:",inline" property:"base"`
	//center.configCenter `yaml:"-"`
	Filter              string                              `yaml:"filter" json:"filter,omitempty" property:"filter"`
	ProxyFactory        string                              `yaml:"proxy_factory" default:"default" json:"proxy_factory,omitempty" property:"proxy_factory"`
	Services            map[string]*service.Config          `yaml:"services" json:"services,omitempty" property:"services"`
	Protocols           map[string]*protocol.ProtocolConfig `yaml:"protocols" json:"protocols,omitempty" property:"protocols"`
	ProtocolConf        interface{}                         `yaml:"protocol_conf" json:"protocol_conf,omitempty" property:"protocol_conf"`
	FilterConf          interface{}                         `yaml:"filter_conf" json:"filter_conf,omitempty" property:"filter_conf"`
	ShutdownConfig      *shutdown.ShutdownConfig            `yaml:"shutdown_conf" json:"shutdown_conf,omitempty" property:"shutdown_conf"`
	ConfigType          map[string]string                   `yaml:"config_type" json:"config_type,omitempty" property:"config_type"`

	Registry   *registry.RegistryConfig            `yaml:"registry" json:"registry,omitempty" property:"registry"`
	Registries map[string]*registry.RegistryConfig `default:"{}" yaml:"registries" json:"registries" property:"registries"`
}

// UnmarshalYAML unmarshals the ProviderConfig by @unmarshal function
func (c *ProviderConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	type plain ProviderConfig
	return unmarshal((*plain)(c))
}

// nolint
func (*ProviderConfig) Prefix() string {
	return constant.ProviderConfigPrefix
}

//// SetProviderConfig sets provider config by @p
//func SetProviderConfig(p ProviderConfig) {
//	config.providerConfig = &p
//}
//
//// ProviderInit loads config file to init provider config
//func ProviderInit(confProFile string) error {
//	if len(confProFile) == 0 {
//		return perrors.Errorf("application configure(provider) file name is nil")
//	}
//	config.providerConfig = &ProviderConfig{}
//	fileStream, err := yaml.UnmarshalYMLConfig(confProFile, config.providerConfig)
//	if err != nil {
//		return perrors.Errorf("unmarshalYmlConfig error %v", perrors.WithStack(err))
//	}
//
//	config.providerConfig.fileStream = bytes.NewBuffer(fileStream)
//	// set method interfaceId & interfaceName
//	for k, v := range config.providerConfig.Services {
//		// set id for reference
//		for _, n := range config.providerConfig.Services[k].Methods {
//			n.InterfaceName = v.InterfaceName
//			n.InterfaceId = k
//		}
//	}
//
//	return nil
//}
//
//func configCenterRefreshProvider() error {
//	// fresh it
//	if config.providerConfig.ConfigCenterConfig != nil {
//		config.providerConfig.fatherConfig = config.providerConfig
//		if err := config.providerConfig.startConfigCenter((*config.providerConfig).BaseConfig); err != nil {
//			return perrors.Errorf("start config center error , error message is {%v}", perrors.WithStack(err))
//		}
//		config.providerConfig.fresh()
//	}
//	return nil
//}