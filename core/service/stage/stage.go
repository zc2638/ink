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

package stage

import (
	"context"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
)

func New() service.Stage {
	return &srv{}
}

type srv struct{}

func (s *srv) List(ctx context.Context, page *v1.Pagination) ([]*v1.Stage, error) {
	db := database.FromContext(ctx)

	var (
		list  []storageV1.Stage
		total int64
	)
	if err := db.Model(&storageV1.Stage{}).Count(&total).Error; err != nil {
		return nil, err
	}
	page.SetTotal(total)

	if err := db.Scopes(page.Scope).Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*v1.Stage, 0, len(list))
	for _, v := range list {
		item, err := v.ToAPI()
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *srv) Info(ctx context.Context, namespace, name string) (*v1.Stage, error) {
	db := database.FromContext(ctx)

	stageS := &storageV1.Stage{Namespace: namespace, Name: name}
	if err := db.Where(stageS).First(stageS).Error; err != nil {
		return nil, err
	}
	return stageS.ToAPI()
}

func (s *srv) Create(ctx context.Context, data *v1.Stage) error {
	db := database.FromContext(ctx)

	var count int64
	stageS := &storageV1.Stage{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(stageS).Model(stageS).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return constant.ErrAlreadyExists
	}

	if err := stageS.FromAPI(data); err != nil {
		return err
	}
	return db.Create(stageS).Error
}

func (s *srv) Update(ctx context.Context, data *v1.Stage) error {
	db := database.FromContext(ctx)
	stageS := &storageV1.Stage{
		Namespace: data.GetNamespace(),
		Name:      data.GetName(),
	}
	if err := db.Where(stageS).First(stageS).Error; err != nil {
		return err
	}
	if err := stageS.FromAPI(data); err != nil {
		return err
	}

	where := &storageV1.Stage{
		Namespace: data.GetNamespace(),
		Name:      data.GetName(),
	}
	return db.Model(where).Where(where).Updates(stageS).Error
}

func (s *srv) Delete(ctx context.Context, namespace, name string) error {
	db := database.FromContext(ctx)

	var count int64
	stageS := &storageV1.Stage{Namespace: namespace, Name: name}
	if err := db.Where(stageS).Model(stageS).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return constant.ErrNoRecord
	}
	return db.Where(stageS).Delete(stageS).Error
}
