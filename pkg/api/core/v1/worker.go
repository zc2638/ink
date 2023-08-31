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

type Worker struct {
	Kind     WorkerKind        `json:"kind,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
	Platform *Platform         `json:"platform,omitempty"`
}

// Platform defines the target platform.
type Platform struct {
	OS   string `json:"os,omitempty"`
	Arch string `json:"arch,omitempty"`
}

type WorkerKind string

func (s WorkerKind) String() string {
	return string(s)
}

const (
	WorkerKindHost       WorkerKind = "host"
	WorkerKindDocker     WorkerKind = "docker"
	WorkerKindKubernetes WorkerKind = "kubernetes"
	WorkerKindSSH        WorkerKind = "ssh"
)
