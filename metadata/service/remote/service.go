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

package remote

import (
	"sync"

	"go.uber.org/atomic"
)

import (
	"github.com/apache/dubbo-go/common"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/apache/dubbo-go/config"
	"github.com/apache/dubbo-go/metadata/definition"
	"github.com/apache/dubbo-go/metadata/identifier"
	"github.com/apache/dubbo-go/metadata/report/delegate"
	"github.com/apache/dubbo-go/metadata/service"
	"github.com/apache/dubbo-go/metadata/service/inmemory"
)

// version will be used by Version func
const version = "1.0.0"

// MetadataService is a implement of metadata service which will delegate the remote metadata report
// This is singleton
type MetadataService struct {
	service.BaseMetadataService
	inMemoryMetadataService *inmemory.MetadataService
	exportedRevision        atomic.String
	subscribedRevision      atomic.String
	delegateReport          *delegate.MetadataReport
}

var (
	metadataServiceOnce     sync.Once
	metadataServiceInstance *MetadataService
)

// NewMetadataService will create a new remote MetadataService instance
func NewMetadataService() (*MetadataService, error) {
	var err error
	metadataServiceOnce.Do(func() {
		var mr *delegate.MetadataReport
		mr, err = delegate.NewMetadataReport()
		if err != nil {
			return
		}
		metadataServiceInstance = &MetadataService{
			inMemoryMetadataService: inmemory.NewMetadataService(),
			delegateReport:          mr,
		}
	})
	return metadataServiceInstance, err
}

// setInMemoryMetadataService will replace the in memory metadata service by the specific param
func (mts *MetadataService) setInMemoryMetadataService(metadata *inmemory.MetadataService) {
	mts.inMemoryMetadataService = metadata
}

// ExportURL will be implemented by in memory service
func (mts *MetadataService) ExportURL(url common.URL) (bool, error) {
	return mts.inMemoryMetadataService.ExportURL(url)
}

// UnexportURL
func (mts *MetadataService) UnexportURL(url common.URL) error {
	smi := identifier.NewServiceMetadataIdentifier(url)
	smi.Revision = mts.exportedRevision.Load()
	return mts.delegateReport.RemoveServiceMetadata(smi)
}

// SubscribeURL will be implemented by in memory service
func (mts *MetadataService) SubscribeURL(url common.URL) (bool, error) {
	return mts.inMemoryMetadataService.SubscribeURL(url)
}

// UnsubscribeURL will be implemented by in memory service
func (mts *MetadataService) UnsubscribeURL(url common.URL) error {
	return mts.UnsubscribeURL(url)
}

// PublishServiceDefinition will call remote metadata's StoreProviderMetadata to store url info and service definition
func (mts *MetadataService) PublishServiceDefinition(url common.URL) error {
	interfaceName := url.GetParam(constant.INTERFACE_KEY, "")
	isGeneric := url.GetParamBool(constant.GENERIC_KEY, false)
	if len(interfaceName) > 0 && !isGeneric {
		sv := common.ServiceMap.GetService(url.Protocol, url.GetParam(constant.BEAN_NAME_KEY, url.Service()))
		sd := definition.BuildServiceDefinition(*sv, url)
		id := &identifier.MetadataIdentifier{
			BaseMetadataIdentifier: identifier.BaseMetadataIdentifier{
				ServiceInterface: interfaceName,
				Version:          url.GetParam(constant.VERSION_KEY, ""),
				// Group:            url.GetParam(constant.GROUP_KEY, constant.SERVICE_DISCOVERY_DEFAULT_GROUP),
				Group: url.GetParam(constant.GROUP_KEY, "test"),
			},
		}
		mts.delegateReport.StoreProviderMetadata(id, sd)
		return nil
	}
	logger.Errorf("publishProvider interfaceName is empty . providerUrl:%v ", url)
	return nil
}

// GetExportedURLs will be implemented by in memory service
func (mts *MetadataService) GetExportedURLs(serviceInterface string, group string, version string, protocol string) ([]common.URL, error) {
	return mts.inMemoryMetadataService.GetExportedURLs(serviceInterface, group, version, protocol)
}

// GetSubscribedURLs will be implemented by in memory service
func (mts *MetadataService) GetSubscribedURLs() ([]common.URL, error) {
	return mts.inMemoryMetadataService.GetSubscribedURLs()
}

// GetServiceDefinition will be implemented by in memory service
func (mts *MetadataService) GetServiceDefinition(interfaceName string, group string, version string) (string, error) {
	return mts.inMemoryMetadataService.GetServiceDefinition(interfaceName, group, version)
}

// GetServiceDefinitionByServiceKey will be implemented by in memory service
func (mts *MetadataService) GetServiceDefinitionByServiceKey(serviceKey string) (string, error) {
	return mts.inMemoryMetadataService.GetServiceDefinitionByServiceKey(serviceKey)
}

// RefreshMetadata will refresh the exported & subscribed metadata to remote metadata report from the inmemory metadata service
func (mts *MetadataService) RefreshMetadata(exportedRevision string, subscribedRevision string) bool {
	result := true
	if len(exportedRevision) != 0 && exportedRevision != mts.exportedRevision.Load() {
		mts.exportedRevision.Store(exportedRevision)
		urls, err := mts.inMemoryMetadataService.GetExportedURLs(constant.ANY_VALUE, "", "", "")
		if err != nil {
			logger.Errorf("Error occur when execute remote.MetadataService.RefreshMetadata, error message is %+v", err)
			result = false
		}
		logger.Infof("urls length = %v", len(urls))
		for _, u := range urls {
			id := identifier.NewServiceMetadataIdentifier(u)
			id.Revision = mts.exportedRevision.Load()
			if err := mts.delegateReport.SaveServiceMetadata(id, u); err != nil {
				logger.Errorf("Error occur when execute remote.MetadataService.RefreshMetadata, error message is %+v", err)
				result = false
			}
		}
	}

	if len(subscribedRevision) != 0 && subscribedRevision != mts.subscribedRevision.Load() {
		mts.subscribedRevision.Store(subscribedRevision)
		urls, err := mts.inMemoryMetadataService.GetSubscribedURLs()
		if err != nil {
			logger.Errorf("Error occur when execute remote.MetadataService.RefreshMetadata, error message is %v+", err)
			result = false
		}
		if urls != nil && len(urls) > 0 {
			id := &identifier.SubscriberMetadataIdentifier{
				MetadataIdentifier: identifier.MetadataIdentifier{
					Application: config.GetApplicationConfig().Name,
				},
				Revision: subscribedRevision,
			}
			if err := mts.delegateReport.SaveSubscribedData(id, urls); err != nil {
				logger.Errorf("Error occur when execute remote.MetadataService.RefreshMetadata, error message is %+v", err)
				result = false
			}
		}
	}
	return result
}

// Version will return the remote service version
func (MetadataService) Version() string {
	return version
}
