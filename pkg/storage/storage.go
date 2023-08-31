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

package storage

import (
	"context"
	"fmt"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

type Config struct {
	File *ConfigFile `json:"file,omitempty"`
}

func New(cfg Config) (Interface, error) {
	if cfg.File != nil {
		return NewFile(*cfg.File)
	}
	return nil, fmt.Errorf("unknown driver")
}

type Selector map[string]string

type Interface interface {
	List(ctx context.Context, meta v1.Metadata, selector Selector) ([]v1.Object, error)
	Info(ctx context.Context, meta v1.Metadata) (v1.Object, error)
	Create(ctx context.Context, meta v1.Metadata, object v1.Object) error
	Update(ctx context.Context, meta v1.Metadata, object v1.Object) error
	Delete(ctx context.Context, meta v1.Metadata) error
}
