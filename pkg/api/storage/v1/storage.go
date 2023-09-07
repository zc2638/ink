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

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

type Model struct {
	ID        uint64 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m *Model) SetID(id uint64) {
	m.ID = id
}

func (m *Model) GetID() uint64 {
	return m.ID
}

type Secret struct {
	Model

	Namespace string
	Name      string
	Data      string
}

func (s *Secret) TableName() string {
	return "secrets"
}

func (s *Secret) FromAPI(in *v1.Secret) error {
	s.Namespace = in.GetNamespace()
	s.Name = in.GetName()
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	s.Data = string(b)
	return nil
}

func (s *Secret) ToAPI() (*v1.Secret, error) {
	var out v1.Secret
	if err := json.Unmarshal([]byte(s.Data), &out); err != nil {
		return nil, err
	}
	out.Namespace = s.Namespace
	out.Name = s.Name
	out.Creation = s.CreatedAt
	return &out, nil
}

type Workflow struct {
	Model

	Namespace string
	Name      string
	Data      string
}

func (s *Workflow) TableName() string {
	return "workflows"
}

func (s *Workflow) FromAPI(in *v1.Workflow) error {
	s.Namespace = in.GetNamespace()
	s.Name = in.GetName()
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	s.Data = string(b)
	return nil
}

func (s *Workflow) ToAPI() (*v1.Workflow, error) {
	var out v1.Workflow
	if err := json.Unmarshal([]byte(s.Data), &out); err != nil {
		return nil, err
	}
	out.Namespace = s.Namespace
	out.Name = s.Name
	out.Creation = s.CreatedAt
	return &out, nil
}

type Box struct {
	Model

	Namespace string
	Name      string
	Enabled   bool
	Data      string
}

func (s *Box) TableName() string {
	return "boxes"
}

func (s *Box) FromAPI(in *v1.Box) error {
	s.Namespace = in.GetNamespace()
	s.Name = in.GetName()
	labels := in.GetLabels()
	status, ok := labels[v1.LabelStatus]
	if !ok {
		status = v1.StatusEnable
	}
	if status != v1.StatusEnable && status != v1.StatusDisable {
		status = v1.StatusEnable
	}
	s.Enabled = status == v1.StatusEnable
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	s.Data = string(b)
	return nil
}

func (s *Box) ToAPI() (*v1.Box, error) {
	var out v1.Box
	if err := json.Unmarshal([]byte(s.Data), &out); err != nil {
		return nil, err
	}
	out.ID = s.ID
	out.Namespace = s.Namespace
	out.Name = s.Name
	out.Creation = s.CreatedAt
	if out.Labels == nil {
		out.Labels = make(map[string]string)
	}
	if s.Enabled {
		out.Labels[v1.LabelStatus] = v1.StatusEnable
	} else {
		out.Labels[v1.LabelStatus] = v1.StatusDisable
	}
	return &out, nil
}

type Build struct {
	Model

	BoxID   uint64 `gorm:"column:box_id"`
	Number  uint64
	Phase   string
	Title   string
	Started int64
	Stopped int64
}

func (s *Build) TableName() string {
	return "builds"
}

func (s *Build) FromAPI(in *v1.Build) {
	s.ID = in.ID
	s.BoxID = in.BoxID
	s.Number = in.Number
	s.Phase = in.Phase.String()
	s.Started = in.Started
	s.Stopped = in.Stopped
	s.Title = in.Title
}

func (s *Build) ToAPI() *v1.Build {
	return &v1.Build{
		ID:      s.ID,
		BoxID:   s.BoxID,
		Number:  s.Number,
		Phase:   v1.Phase(s.Phase),
		Title:   s.Title,
		Started: s.Started,
		Stopped: s.Stopped,
	}
}

type Stage struct {
	Model

	BoxID      uint64
	BuildID    uint64
	Number     uint64
	Phase      string
	Name       string
	WorkerName string
	Worker     string
	Started    int64
	Stopped    int64
	Error      string
}

func (s *Stage) TableName() string {
	return "stages"
}

func (s *Stage) FromAPI(in *v1.Stage) error {
	b, err := json.Marshal(in.Worker)
	if err != nil {
		return err
	}
	s.Worker = string(b)
	s.ID = in.ID
	s.BoxID = in.BoxID
	s.BuildID = in.BuildID
	s.Number = in.Number
	s.Phase = in.Phase.String()
	s.Name = in.Name
	s.WorkerName = in.WorkerName
	s.Started = in.Started
	s.Stopped = in.Stopped
	s.Error = in.Error
	return nil
}

func (s *Stage) ToAPI() (*v1.Stage, error) {
	result := &v1.Stage{
		ID:      s.ID,
		BoxID:   s.BoxID,
		BuildID: s.BuildID,
		Number:  s.Number,
		Phase:   v1.Phase(s.Phase),
		Name:    s.Name,
		Started: s.Started,
		Stopped: s.Stopped,
		Error:   s.Error,
	}
	if err := json.Unmarshal([]byte(s.Worker), &result.Worker); err != nil {
		return nil, err
	}
	return result, nil
}

type Step struct {
	Model

	StageID  uint64
	Number   uint64
	Phase    string
	Name     string
	Started  int64
	Stopped  int64
	ExitCode int
	Error    string
}

func (s *Step) TableName() string {
	return "steps"
}

func (s *Step) FromAPI(in *v1.Step) {
	s.ID = in.ID
	s.StageID = in.StageID
	s.Number = in.Number
	s.Phase = in.Phase.String()
	s.Name = in.Name
	s.Started = in.Started
	s.Stopped = in.Stopped
	s.ExitCode = in.ExitCode
	s.Error = in.Error
}

func (s *Step) ToAPI() *v1.Step {
	return &v1.Step{
		ID:       s.ID,
		StageID:  s.StageID,
		Number:   s.Number,
		Phase:    v1.Phase(s.Phase),
		Name:     s.Name,
		Started:  s.Started,
		Stopped:  s.Stopped,
		ExitCode: s.ExitCode,
		Error:    s.Error,
	}
}

type Log struct {
	Model

	Data []byte
}

func (s *Log) TableName() string {
	return "logs"
}
