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

package v1

import (
	"net/http"
	"net/url"
	"strconv"

	"gorm.io/gorm"
)

func GetPagination(r *http.Request) *Pagination {
	page, _ := strconv.Atoi(
		r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(
		r.URL.Query().Get("size"))
	return &Pagination{
		Page: page,
		Size: size,
	}
}

type Pagination struct {
	Page  int   `json:"page" desc:"page number"`
	Size  int   `json:"size" desc:"page size"`
	Total int64 `json:"total" desc:"total number"`
}

func (o *Pagination) complete() {
	if o.Page < 1 {
		o.Page = 1
	}
	if o.Size < 1 {
		o.Size = 10
	}
	if o.Size > 100 {
		o.Size = 100
	}
}

func (o *Pagination) SetTotal(total int64) *Pagination {
	o.Total = total
	return o
}

func (o *Pagination) Scope(db *gorm.DB) *gorm.DB {
	o.complete()
	offset := (o.Page - 1) * o.Size
	return db.Limit(o.Size).Offset(offset)
}

func (o *Pagination) List(list any) map[string]any {
	if list == nil {
		list = []struct{}{}
	}
	return map[string]any{
		"page":  o.Page,
		"size":  o.Size,
		"total": o.Total,
		"items": list,
	}
}

func (o *Pagination) ToValues() url.Values {
	o.complete()
	vs := make(url.Values)
	vs.Set("page", strconv.Itoa(o.Page))
	vs.Set("size", strconv.Itoa(o.Size))
	return vs
}
