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

package box

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
)

func validateResources(db *gorm.DB, box *v1.Box) error {
	workflows := make([]*v1.Workflow, 0, len(box.Resources))
	for _, rv := range box.Resources {
		if rv.Kind == "" {
			rv.Kind = v1.KindWorkflow
		}
		if rv.Kind != v1.KindWorkflow {
			continue
		}

		sd := &storageV1.Workflow{
			Namespace: box.GetNamespace(),
			Name:      rv.Name,
		}
		if err := db.Where(sd).First(sd).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("workflow not found: %s", sd.Name)
			}
			return err
		}
		workflow, err := sd.ToAPI()
		if err != nil {
			return err
		}
		workflows = append(workflows, workflow)
	}

	if len(workflows) == 0 {
		return errors.New("resources did not find a workflow")
	}
	return box.Validate(workflows)
}
