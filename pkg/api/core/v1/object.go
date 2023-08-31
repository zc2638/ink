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
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

type Object interface {
	GetNamespace() string
	SetNamespace(namespace string)
	GetKind() string
	SetKind(kind string)
	GetName() string
	SetName(name string)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)

	GetCreationTimestamp() time.Time
	SetCreationTimestamp(timestamp time.Time)
	GetDeletionTimestamp() *time.Time
	SetDeletionTimestamp(timestamp *time.Time)
}

type UnstructuredObject struct {
	Object map[string]any
}

func (o *UnstructuredObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Object)
}

func (o *UnstructuredObject) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &o.Object)
}

func (o *UnstructuredObject) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(o.Object)
}

func (o *UnstructuredObject) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&o.Object)
}

func (o *UnstructuredObject) ToObject(v Object) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (o *UnstructuredObject) GetNamespace() string {
	ns := getString(o.Object, "namespace")
	if ns == "" {
		ns = DefaultNamespace
	}
	return ns
}

func (o *UnstructuredObject) SetNamespace(namespace string) {
	SetValueToMap(o.Object, namespace, "namespace")
}

func (o *UnstructuredObject) GetKind() string {
	return getString(o.Object, "kind")
}

func (o *UnstructuredObject) SetKind(kind string) {
	SetValueToMap(o.Object, kind, "kind")
}

func (o *UnstructuredObject) GetName() string {
	return getString(o.Object, "name")
}

func (o *UnstructuredObject) SetName(name string) {
	SetValueToMap(o.Object, name, "name")
}

func (o *UnstructuredObject) GetLabels() map[string]string {
	v, ok := GetValueFromMap(o.Object, "labels")
	if !ok {
		return nil
	}
	labels, ok := v.(map[string]string)
	if !ok {
		return nil
	}
	return labels
}

func (o *UnstructuredObject) SetLabels(labels map[string]string) {
	SetValueToMap(o.Object, labels, "labels")
}

func (o *UnstructuredObject) GetCreationTimestamp() time.Time {
	t := getTime(o.Object, "creationTimestamp")
	if t == nil {
		return time.Time{}
	}
	return *t
}

func (o *UnstructuredObject) SetCreationTimestamp(timestamp time.Time) {
	SetValueToMap(o.Object, timestamp, "creationTimestamp")
}

func (o *UnstructuredObject) GetDeletionTimestamp() *time.Time {
	return getTime(o.Object, "deletionTimestamp")
}

func (o *UnstructuredObject) SetDeletionTimestamp(timestamp *time.Time) {
	SetValueToMap(o.Object, timestamp, "deletionTimestamp")
}

func getString(obj map[string]any, fromPath ...string) string {
	v, ok := GetValueFromMap(obj, fromPath...)
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func getTime(obj map[string]any, fromPath ...string) *time.Time {
	s := getString(obj, fromPath...)
	if len(s) == 0 {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func GetValueFromMap(obj map[string]any, fromPath ...string) (any, bool) {
	var (
		fromVal any
		ok      bool
		isObj   = true
	)
	for _, fp := range fromPath {
		if !isObj {
			return nil, false
		}

		fromVal, ok = obj[fp]
		if !ok {
			return nil, false
		}
		obj, isObj = fromVal.(map[string]any)
	}
	return fromVal, true
}

func SetValueToMap(obj map[string]any, toVal any, toPath ...string) {
	lastIndex := len(toPath) - 1
	for i, tp := range toPath {
		if lastIndex == i {
			obj[tp] = toVal
			break
		}

		val, ok := obj[tp]
		if !ok {
			val = make(map[string]any)
			obj[tp] = val
		}
		if _, ok := val.(map[string]any); !ok {
			val = make(map[string]any)
			obj[tp] = val
		} else if val == nil {
			val = make(map[string]any)
			obj[tp] = val
		}
		obj = val.(map[string]any)
	}
}
