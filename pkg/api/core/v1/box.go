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
	"fmt"

	"github.com/99nil/gopkg/cycle"
	"github.com/99nil/gopkg/sets"

	"github.com/zc2638/ink/pkg/selector"
)

// Box defines a collection of stage executions.
type Box struct {
	Metadata `yaml:",inline"`

	Resources []BoxResource     `json:"resources" yaml:"resources"`
	Settings  map[string]string `json:"settings,omitempty" yaml:"settings,omitempty"`
}

func (b *Box) GetSelectors(kind string, settings map[string]string) (names []string, selectors []*selector.Selector) {
	nameSet := sets.New[string]()
	for _, v := range b.Resources {
		if v.Kind != kind {
			continue
		}
		if v.Selector != nil {
			if ok := v.Selector.Match(settings); !ok {
				continue
			}
		}

		nameSet.Add(v.Name)
		if v.LabelSelector != nil {
			selectors = append(selectors, v.LabelSelector)
		}
	}
	names = nameSet.List()
	return
}

func (b *Box) Validate(workflows []*Workflow) error {
	graph := cycle.New()
	for _, s := range workflows {
		graph.Add(s.Name, s.Spec.DependsOn...)
	}
	if graph.DetectCycles() {
		return errors.New("dependency cycle detected in workflows")
	}

	for index, rv := range b.Resources {
		if rv.Name == "" && rv.Selector == nil && rv.LabelSelector == nil {
			return fmt.Errorf("invalid resource at index: %d", index)
		}
		if rv.Selector != nil {
			if err := rv.Selector.Validate(); err != nil {
				return err
			}
		}
		if rv.LabelSelector != nil {
			if err := rv.LabelSelector.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

type BoxResource struct {
	Kind          string             `json:"kind" yaml:"kind"`
	Name          string             `json:"name,omitempty" yaml:"name,omitempty"`
	Selector      *selector.Selector `json:"selector,omitempty" yaml:"selector,omitempty"`
	LabelSelector *selector.Selector `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}
