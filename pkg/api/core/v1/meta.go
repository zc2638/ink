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
	"fmt"
	"time"
)

const (
	DefaultNamespace = "default"
	AllNamespace     = ""
)

const (
	KindBox      = "Box"
	KindWorkflow = "Workflow"
	KindSecret   = "Secret"
)

const LabelStatus = "ink.io/status"

const (
	StatusEnable  = "enable"
	StatusDisable = "disable"
)

func NewMetadata(namespace, name string) Metadata {
	return Metadata{Namespace: namespace, Name: name}
}

func GetMetadata(object Object) Metadata {
	return Metadata{
		Kind:      object.GetKind(),
		Name:      object.GetName(),
		Namespace: object.GetNamespace(),
		Labels:    object.GetLabels(),
		Creation:  object.GetCreationTimestamp(),
		Deletion:  object.GetDeletionTimestamp(),
	}
}

type Metadata struct {
	Kind      string            `json:"kind" yaml:"kind"`
	Name      string            `json:"name" yaml:"name"`
	Namespace string            `json:"namespace" yaml:"namespace"`
	Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	ID       uint64     `json:"id,omitempty" yaml:"id,omitempty"`
	Creation time.Time  `json:"creation,omitempty" yaml:"creation,omitempty"`
	Deletion *time.Time `json:"deletion,omitempty" yaml:"deletion,omitempty"`
}

func (m *Metadata) String() string {
	return fmt.Sprintf("namespace=%s, kind=%s, name=%s", m.GetNamespace(), m.GetKind(), m.GetName())
}

func (m *Metadata) GetNamespace() string {
	if m.Namespace == "" {
		return DefaultNamespace
	}
	return m.Namespace
}

func (m *Metadata) SetNamespace(namespace string)             { m.Namespace = namespace }
func (m *Metadata) GetKind() string                           { return m.Kind }
func (m *Metadata) SetKind(kind string)                       { m.Kind = kind }
func (m *Metadata) GetName() string                           { return m.Name }
func (m *Metadata) SetName(name string)                       { m.Name = name }
func (m *Metadata) GetLabels() map[string]string              { return m.Labels }
func (m *Metadata) SetLabels(labels map[string]string)        { m.Labels = labels }
func (m *Metadata) GetCreationTimestamp() time.Time           { return m.Creation }
func (m *Metadata) SetCreationTimestamp(timestamp time.Time)  { m.Creation = timestamp }
func (m *Metadata) GetDeletionTimestamp() *time.Time          { return m.Deletion }
func (m *Metadata) SetDeletionTimestamp(timestamp *time.Time) { m.Deletion = timestamp }
