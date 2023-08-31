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
	stages := make([]*v1.Stage, 0, len(box.Resources))
	for _, rv := range box.Resources {
		if rv.Kind != v1.KindStage {
			continue
		}

		stageS := &storageV1.Stage{
			Namespace: box.GetNamespace(),
			Name:      rv.Name,
		}
		if err := db.Where(stageS).First(stageS).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("stage not found: %s", stageS.Name)
			}
			return err
		}
		stage, err := stageS.ToAPI()
		if err != nil {
			return err
		}
		stages = append(stages, stage)
	}

	if len(stages) == 0 {
		return errors.New("resources did not find a stage")
	}
	return box.Validate(stages)
}
