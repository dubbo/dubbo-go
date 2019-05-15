package config

import (
	"github.com/dubbo/go-for-apache-dubbo/common/extension"
	"github.com/stretchr/testify/assert"
	"testing"
)

func doinit() {
	providerConfig = &ProviderConfig{
		ApplicationConfig: ApplicationConfig{
			Organization: "dubbo_org",
			Name:         "dubbo",
			Module:       "module",
			Version:      "2.6.0",
			Owner:        "dubbo",
			Environment:  "test"},
		Registries: []RegistryConfig{
			{
				Id:         "shanghai_reg1",
				Type:       "mock",
				TimeoutStr: "2s",
				Group:      "shanghai_idc",
				Address:    "127.0.0.1:2181",
				Username:   "user1",
				Password:   "pwd1",
			},
			{
				Id:         "shanghai_reg2",
				Type:       "mock",
				TimeoutStr: "2s",
				Group:      "shanghai_idc",
				Address:    "127.0.0.2:2181",
				Username:   "user1",
				Password:   "pwd1",
			},
			{
				Id:         "hangzhou_reg1",
				Type:       "mock",
				TimeoutStr: "2s",
				Group:      "hangzhou_idc",
				Address:    "127.0.0.3:2181",
				Username:   "user1",
				Password:   "pwd1",
			},
			{
				Id:         "hangzhou_reg2",
				Type:       "mock",
				TimeoutStr: "2s",
				Group:      "hangzhou_idc",
				Address:    "127.0.0.4:2181",
				Username:   "user1",
				Password:   "pwd1",
			},
		},
		Services: []ServiceConfig{
			{
				InterfaceName: "testInterface",
				Protocol:      "mock",
				Registries:    []ConfigRegistry{"shanghai_reg1", "shanghai_reg2", "hangzhou_reg1", "hangzhou_reg2"},
				Cluster:       "failover",
				Loadbalance:   "random",
				Retries:       3,
				Group:         "huadong_idc",
				Version:       "1.0.0",
				Methods: []struct {
					Name        string `yaml:"name"  json:"name,omitempty"`
					Retries     int64  `yaml:"retries"  json:"retries,omitempty"`
					Loadbalance string `yaml:"loadbalance"  json:"loadbalance,omitempty"`
					Weight      int64  `yaml:"weight"  json:"weight,omitempty"`
				}{
					{
						Name:        "GetUser",
						Retries:     2,
						Loadbalance: "random",
						Weight:      200,
					},
					{
						Name:        "GetUser1",
						Retries:     2,
						Loadbalance: "random",
						Weight:      200,
					},
				},
			},
		},
		Protocols: []ProtocolConfig{
			{
				Name:        "mock",
				Ip:          "127.0.0.1",
				Port:        "20000",
				ContextPath: "/xxx",
			},
		},
	}
}
func Test_Export(t *testing.T) {
	doinit()
	extension.SetProtocol("registry", GetProtocol)

	for _, service := range providerConfig.Services {
		service.Implement(&MockService{})
		service.Export()
		assert.Condition(t, func() bool {
			return len(service.exporters) > 0
		})
	}
	providerConfig = nil
}
