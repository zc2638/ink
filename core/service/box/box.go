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

package box

import (
	"context"
	"reflect"

	"gorm.io/gorm"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/service"
	"github.com/zc2638/ink/core/service/common"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
)

func New() service.Box {
	return &srv{}
}

type srv struct{}

func (s *srv) List(ctx context.Context, namespace string, opt v1.ListOption) ([]*v1.Box, error) {
	db := database.FromContext(ctx)

	labels := opt.Labels()
	if len(labels) > 0 {
		names, err := common.SelectNamesByLabels(ctx, v1.KindBox, namespace, labels)
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, nil
		}
		db = db.Where("name in (?)", names)
	}
	if len(namespace) > 0 {
		db = db.Where("namespace = ?", namespace)
	}
	if err := db.Model(&storageV1.Box{}).Count(&opt.Pagination.Total).Error; err != nil {
		return nil, err
	}
	var list []storageV1.Box
	if err := db.Scopes(opt.Pagination.Scope).Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*v1.Box, 0, len(list))
	for _, v := range list {
		item, err := v.ToAPI()
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *srv) Info(ctx context.Context, namespace, name string) (*v1.Box, error) {
	db := database.FromContext(ctx)

	info := &storageV1.Box{Namespace: namespace, Name: name}
	if err := db.Where(info).First(info).Error; err != nil {
		return nil, err
	}
	out, err := info.ToAPI()
	if err != nil {
		return nil, err
	}

	if err := db.Model(&storageV1.Build{}).Where(&storageV1.Build{BoxID: info.ID}).Count(&out.Status.Builds).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *srv) Create(ctx context.Context, data *v1.Box) error {
	db := database.FromContext(ctx)
	if err := validateResources(db, data); err != nil {
		return err
	}

	var count int64
	sd := &storageV1.Box{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(sd).Model(sd).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return constant.ErrAlreadyExists
	}

	if err := sd.FromAPI(data); err != nil {
		return err
	}
	labels := common.ConvertLabels(v1.KindBox, sd.Namespace, sd.Name, data.Labels)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := db.Create(sd).Error; err != nil {
			return err
		}
		if len(labels) > 0 {
			if err := db.CreateInBatches(labels, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *srv) Update(ctx context.Context, data *v1.Box) error {
	db := database.FromContext(ctx)

	if err := validateResources(db, data); err != nil {
		return err
	}

	sd := &storageV1.Box{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(sd).First(sd).Error; err != nil {
		return err
	}
	origin, err := sd.ToAPI()
	if err != nil {
		return err
	}
	if err := sd.FromAPI(data); err != nil {
		return err
	}

	var labels []storageV1.Label
	labelChanged := !reflect.DeepEqual(origin.Labels, data.Labels)
	if labelChanged {
		labels = common.ConvertLabels(v1.KindBox, sd.Namespace, sd.Name, data.Labels)
	}
	where := &storageV1.Box{Namespace: sd.Namespace, Name: sd.Name}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := db.Model(where).Where(where).Updates(sd).Error; err != nil {
			return err
		}
		if labelChanged {
			if err := db.Where(&storageV1.Label{
				Namespace: where.Namespace,
				Name:      where.Name,
				Kind:      v1.KindBox,
			}).Delete(&storageV1.Label{}).Error; err != nil {
				return err
			}
		}
		if len(labels) > 0 {
			if err := db.CreateInBatches(labels, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *srv) Delete(ctx context.Context, namespace, name string) error {
	db := database.FromContext(ctx)

	var count int64
	sd := &storageV1.Box{Namespace: namespace, Name: name}
	if err := db.Where(sd).Model(sd).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return constant.ErrNoRecord
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := db.Where(sd).Delete(sd).Error; err != nil {
			return err
		}
		return db.Where(&storageV1.Label{
			Namespace: sd.Namespace,
			Name:      sd.Name,
			Kind:      v1.KindBox,
		}).Delete(&storageV1.Label{}).Error
	})
}
