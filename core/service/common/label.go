// Copyright © 2023 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"
	"fmt"

	"github.com/99nil/gopkg/sets"

	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
)

func SelectNamesByLabels(ctx context.Context, kind, namespace string, labels map[string]string) ([]string, error) {
	if len(labels) == 0 {
		return nil, nil
	}

	db := database.FromContext(ctx)
	db = db.Where(&storageV1.Label{Namespace: namespace, Kind: kind})

	var start bool
	nameSet := sets.New[string]()
	for k, v := range labels {
		var selectNames []string
		if err := db.Where("key = ?", k).Where("value = ?", v).Pluck("name", &selectNames).Error; err != nil {
			return nil, fmt.Errorf("select label(%s=%s) failed: %v", k, v, err)
		}

		if !start {
			start = true
			nameSet.Add(selectNames...)
			continue
		}
		for _, sn := range selectNames {
			if !nameSet.Has(sn) {
				nameSet.Remove(sn)
			}
		}
		if nameSet.Len() == 0 {
			return nil, nil
		}
	}
	return nameSet.List(), nil
}
