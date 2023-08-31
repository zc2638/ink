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
	"errors"

	"github.com/99nil/gopkg/cycle"
)

// Box defines a collection of stage executions.
type Box struct {
	Metadata `yaml:",inline"`

	Resources []BoxResource `json:"resources" yaml:"resources"`
}

func (b *Box) Validate(stages []*Stage) error {
	graph := cycle.New()
	for _, s := range stages {
		graph.Add(s.Name, s.Spec.DependsOn...)
	}
	if graph.DetectCycles() {
		return errors.New("dependency cycle detected in Stage")
	}
	return nil
}

type BoxResource struct {
	Kind string `json:"kind" yaml:"kind"`
	Name string `json:"name" yaml:"name"`
}
