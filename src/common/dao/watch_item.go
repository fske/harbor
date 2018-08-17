// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"time"

	"github.com/fske/harbor/src/common/models"
)

// DefaultDatabaseWatchItemDAO is an instance of DatabaseWatchItemDAO
var DefaultDatabaseWatchItemDAO WatchItemDAO = &DatabaseWatchItemDAO{}

// WatchItemDAO defines operations about WatchItem
type WatchItemDAO interface {
	Add(*models.WatchItem) (int64, error)
	DeleteByPolicyID(int64) error
	Get(namespace, operation string) ([]models.WatchItem, error)
}

// DatabaseWatchItemDAO implements interface WatchItemDAO for database
type DatabaseWatchItemDAO struct{}

// Add a WatchItem
func (d *DatabaseWatchItemDAO) Add(item *models.WatchItem) (int64, error) {
	now := time.Now()
	item.CreationTime = now
	item.UpdateTime = now
	return GetOrmer().Insert(item)
}

// DeleteByPolicyID deletes the WatchItem specified by policy ID
func (d *DatabaseWatchItemDAO) DeleteByPolicyID(policyID int64) error {
	_, err := GetOrmer().QueryTable(&models.WatchItem{}).Filter("PolicyID", policyID).Delete()
	return err
}

// Get returns WatchItem list according to the namespace and operation
func (d *DatabaseWatchItemDAO) Get(namespace, operation string) ([]models.WatchItem, error) {
	qs := GetOrmer().QueryTable(&models.WatchItem{}).Filter("Namespace", namespace)
	if operation == "push" {
		qs = qs.Filter("OnPush", 1)
	} else if operation == "delete" {
		qs = qs.Filter("OnDeletion", 1)
	}

	items := []models.WatchItem{}
	_, err := qs.All(&items)
	return items, err
}
