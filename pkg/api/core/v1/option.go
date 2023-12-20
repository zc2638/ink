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
	"net/url"
	"strings"
)

type ListOption struct {
	Pagination    Pagination
	LabelSelector string
}

func (o *ListOption) ToValues() url.Values {
	result := url.Values{}
	for k, v := range o.Pagination.ToValues() {
		result[k] = v
	}
	if len(o.LabelSelector) > 0 {
		result.Set("labelSelector", o.LabelSelector)
	}
	return result
}

func (o *ListOption) SetLabels(labels map[string]string) {
	var sb strings.Builder
	for k, v := range labels {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
		sb.WriteString(",")
	}
	s := sb.String()
	s = strings.TrimSuffix(s, ",")
	o.LabelSelector = s
}

func (o *ListOption) Labels() map[string]string {
	selector := strings.TrimSpace(o.LabelSelector)
	if len(selector) == 0 {
		return nil
	}

	labels := make(map[string]string)
	parts := strings.Split(o.LabelSelector, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		labels[kv[0]] = kv[1]
	}
	return labels
}
