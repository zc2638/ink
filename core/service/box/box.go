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

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
)

func New() service.Box {
	return &srv{}
}

type srv struct{}

func (s *srv) List(ctx context.Context, page *v1.Pagination) ([]*v1.Box, error) {
	db := database.FromContext(ctx)

	var (
		list  []storageV1.Box
		total int64
	)
	if err := db.Model(&storageV1.Box{}).Count(&total).Error; err != nil {
		return nil, err
	}
	page.SetTotal(total)
	if err := db.Scopes(page.Scope).Find(&list).Error; err != nil {
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
	return info.ToAPI()
}

func (s *srv) Create(ctx context.Context, data *v1.Box) error {
	db := database.FromContext(ctx)

	if err := validateResources(db, data); err != nil {
		return err
	}

	var count int64
	boxS := &storageV1.Box{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(boxS).Model(boxS).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return constant.ErrAlreadyExists
	}

	if err := boxS.FromAPI(data); err != nil {
		return err
	}
	return db.Create(boxS).Error
}

func (s *srv) Update(ctx context.Context, data *v1.Box) error {
	db := database.FromContext(ctx)

	if err := validateResources(db, data); err != nil {
		return err
	}

	boxS := &storageV1.Box{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(boxS).First(boxS).Error; err != nil {
		return err
	}
	if err := boxS.FromAPI(data); err != nil {
		return err
	}

	where := &storageV1.Box{
		Namespace: data.GetNamespace(),
		Name:      data.GetName(),
	}
	return db.Model(where).Where(where).Updates(boxS).Error
}

func (s *srv) Delete(ctx context.Context, namespace, name string) error {
	db := database.FromContext(ctx)

	boxS := &storageV1.Box{
		Namespace: namespace,
		Name:      name,
	}

	var count int64
	if err := db.Where(boxS).Model(boxS).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return constant.ErrNoRecord
	}
	return db.Where(boxS).Delete(boxS).Error
}
