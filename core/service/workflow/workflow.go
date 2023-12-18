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

package workflow

import (
	"context"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
)

func New() service.Workflow {
	return &srv{}
}

type srv struct{}

func (s *srv) List(ctx context.Context, namespace string, page *v1.Pagination) ([]*v1.Workflow, error) {
	db := database.FromContext(ctx)

	var (
		list  []storageV1.Workflow
		total int64
	)
	if len(namespace) > 0 {
		db = db.Where("namespace = ?", namespace)
	}
	if err := db.Model(&storageV1.Workflow{}).Count(&total).Error; err != nil {
		return nil, err
	}
	page.SetTotal(total)

	if err := db.Scopes(page.Scope).Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*v1.Workflow, 0, len(list))
	for _, v := range list {
		item, err := v.ToAPI()
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *srv) Info(ctx context.Context, namespace, name string) (*v1.Workflow, error) {
	db := database.FromContext(ctx)

	sd := &storageV1.Workflow{Namespace: namespace, Name: name}
	if err := db.Where(sd).First(sd).Error; err != nil {
		return nil, err
	}
	return sd.ToAPI()
}

func (s *srv) Create(ctx context.Context, data *v1.Workflow) error {
	db := database.FromContext(ctx)

	var count int64
	sd := &storageV1.Workflow{Namespace: data.GetNamespace(), Name: data.GetName()}
	if err := db.Where(sd).Model(sd).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return constant.ErrAlreadyExists
	}

	if err := sd.FromAPI(data); err != nil {
		return err
	}
	return db.Create(sd).Error
}

func (s *srv) Update(ctx context.Context, data *v1.Workflow) error {
	db := database.FromContext(ctx)
	sd := &storageV1.Workflow{
		Namespace: data.GetNamespace(),
		Name:      data.GetName(),
	}
	if err := db.Where(sd).First(sd).Error; err != nil {
		return err
	}
	if err := sd.FromAPI(data); err != nil {
		return err
	}

	where := &storageV1.Workflow{
		Namespace: data.GetNamespace(),
		Name:      data.GetName(),
	}
	return db.Model(where).Where(where).Updates(sd).Error
}

func (s *srv) Delete(ctx context.Context, namespace, name string) error {
	db := database.FromContext(ctx)

	var count int64
	sd := &storageV1.Workflow{Namespace: namespace, Name: name}
	if err := db.Where(sd).Model(sd).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return constant.ErrNoRecord
	}
	return db.Where(sd).Delete(sd).Error
}
