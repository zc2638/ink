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

package service

import (
	"context"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

type (
	Stage interface {
		List(ctx context.Context, page *v1.Pagination) ([]*v1.Stage, error)
		Info(ctx context.Context, namespace, name string) (*v1.Stage, error)
		Create(ctx context.Context, data *v1.Stage) error
		Update(ctx context.Context, data *v1.Stage) error
		Delete(ctx context.Context, namespace, name string) error
	}

	Box interface {
		List(ctx context.Context, page *v1.Pagination) ([]*v1.Box, error)
		Info(ctx context.Context, namespace, name string) (*v1.Box, error)
		Create(ctx context.Context, data *v1.Box) error
		Update(ctx context.Context, data *v1.Box) error
		Delete(ctx context.Context, namespace, name string) error
	}

	Build interface {
		List(ctx context.Context, namespace, name string, page *v1.Pagination) ([]*v1.Build, error)
		Info(ctx context.Context, namespace, name string, number uint64) (*v1.Build, error)
		Create(ctx context.Context, namespace, name string) (uint64, error)
		Cancel(ctx context.Context, namespace, name string, number uint64) error
	}
)
