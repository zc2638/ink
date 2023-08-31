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

type Build struct {
	ID      uint64 `json:"id" yaml:"id"`
	BoxID   uint64 `json:"boxID" yaml:"boxID"`
	Number  uint64 `json:"number" yaml:"number"`
	Phase   Phase  `json:"phase" yaml:"phase"`
	Title   string `json:"title" yaml:"title"`
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	Started int64  `json:"started,omitempty" yaml:"started,omitempty"`
	Stopped int64  `json:"stopped,omitempty" yaml:"stopped,omitempty"`

	Stages []*StageStatus `json:"stages,omitempty" yaml:"stages,omitempty"`
}

type StageStatus struct {
	ID      uint64 `json:"id" yaml:"id"`
	BoxID   uint64 `json:"boxID" yaml:"boxID"`
	BuildID uint64 `json:"buildID" yaml:"buildID"`
	Number  uint64 `json:"number" yaml:"number"`
	Phase   Phase  `json:"phase" yaml:"phase"`
	Name    string `json:"name" yaml:"name"`
	Limit   int    `json:"limit" yaml:"limit"`
	Started int64  `json:"started,omitempty" yaml:"started,omitempty"`
	Stopped int64  `json:"stopped,omitempty" yaml:"stopped,omitempty"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`

	WorkerName string   `json:"workerName,omitempty" yaml:"workerName,omitempty"`
	Worker     Worker   `json:"worker,omitempty" yaml:"worker,omitempty"`
	DependsOn  []string `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`

	Steps []*StepStatus `json:"steps,omitempty" yaml:"steps,omitempty"`
}

type StepStatus struct {
	ID       uint64 `json:"id" yaml:"id"`
	StageID  uint64 `json:"stageID" yaml:"stageID"`
	Number   uint64 `json:"number" yaml:"number"`
	Phase    Phase  `json:"phase" yaml:"phase"`
	Name     string `json:"name" yaml:"name"`
	Started  int64  `json:"started,omitempty" yaml:"started,omitempty"`
	Stopped  int64  `json:"stopped,omitempty" yaml:"stopped,omitempty"`
	ExitCode int    `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`
	Error    string `json:"error,omitempty" yaml:"error,omitempty"`
}
