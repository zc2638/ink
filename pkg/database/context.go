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

package database

import (
	"context"
	"net/http"

	"gorm.io/gorm"
)

type key struct{}

// WithContext returns a new context with the provided db.
func WithContext(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, key{}, db.WithContext(ctx))
}

// FromContext retrieves the current db from the context. If no
// db is available, the nil value is returned.
func FromContext(ctx context.Context) *gorm.DB {
	v := ctx.Value(key{})
	if v == nil {
		return nil
	}
	return v.(*gorm.DB)
}

// FromRequest retrieves the current db from the request. If no
// db is available, the nil value is returned.
func FromRequest(r *http.Request) *gorm.DB {
	return FromContext(r.Context())
}
