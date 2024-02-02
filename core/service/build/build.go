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

package build

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/99nil/gopkg/sets"

	"github.com/zc2638/ink/core/scheduler"
	"gorm.io/gorm"

	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"

	"github.com/zc2638/ink/core/service"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

func New() service.Build {
	return &srv{}
}

type srv struct{}

func (s *srv) List(ctx context.Context, namespace, name string, page *v1.Pagination) ([]*v1.Build, error) {
	db := database.FromContext(ctx)

	boxS := &storageV1.Box{
		Namespace: namespace,
		Name:      name,
	}
	if err := db.Where(boxS).First(boxS).Error; err != nil {
		return nil, err
	}

	var (
		list  []storageV1.Build
		total int64
	)
	buildS := &storageV1.Build{BoxID: boxS.ID}
	db = db.Where(buildS)
	if err := db.Model(buildS).Count(&total).Error; err != nil {
		return nil, err
	}
	page.SetTotal(total)
	if err := db.Scopes(page.Scope).Order("number desc").Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*v1.Build, 0, len(list))
	for _, v := range list {
		item, err := v.ToAPI()
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *srv) Info(ctx context.Context, namespace, name string, number uint64) (*v1.Build, error) {
	db := database.FromContext(ctx)

	boxS := &storageV1.Box{
		Namespace: namespace,
		Name:      name,
	}
	if err := db.Where(boxS).First(boxS).Error; err != nil {
		return nil, err
	}

	buildS := &storageV1.Build{
		BoxID:  boxS.ID,
		Number: number,
	}
	if err := db.Where(buildS).First(buildS).Error; err != nil {
		return nil, err
	}
	build, err := buildS.ToAPI()
	if err != nil {
		return nil, err
	}

	var stageList []storageV1.Stage
	stageS := &storageV1.Stage{
		BoxID:   buildS.BoxID,
		BuildID: buildS.ID,
	}
	if err := db.Where(stageS).Find(&stageList).Error; err != nil {
		return nil, err
	}
	for _, v := range stageList {
		stage, err := v.ToAPI()
		if err != nil {
			return nil, err
		}

		var stepList []storageV1.Step
		stepS := &storageV1.Step{StageID: v.ID}
		if err := db.Where(stepS).Find(&stepList).Error; err != nil {
			return nil, err
		}
		for _, sv := range stepList {
			step := sv.ToAPI()
			stage.Steps = append(stage.Steps, step)
		}
		build.Stages = append(build.Stages, stage)
	}
	return build, nil
}

func (s *srv) Create(ctx context.Context, namespace, name string, settings map[string]string) (uint64, error) {
	db := database.FromContext(ctx)

	boxS := &storageV1.Box{
		Namespace: namespace,
		Name:      name,
	}
	if err := db.Where(boxS).First(boxS).Error; err != nil {
		return 0, err
	}
	box, err := boxS.ToAPI()
	if err != nil {
		return 0, err
	}

	var buildCount int64
	where := &storageV1.Build{BoxID: box.ID}
	if err := db.Where(where).Model(where).Count(&buildCount).Error; err != nil {
		return 0, err
	}

	build := &v1.Build{
		BoxID:    box.ID,
		Number:   uint64(buildCount) + 1,
		Phase:    v1.PhasePending,
		Settings: settings,
	}
	var buildS storageV1.Build
	if err := buildS.FromAPI(build); err != nil {
		return 0, err
	}

	workflowNames, selectors := box.GetSelectors(v1.KindWorkflow, settings)
	if len(workflowNames) == 0 {
		return 0, errors.New("workflow resource not found")
	}

	var workflowList []storageV1.Workflow
	dbW := db.Where("namespace = ?", box.Namespace)
	if !slices.Contains(workflowNames, "") {
		dbW = dbW.Where("name in (?)", workflowNames)
	}
	if err := dbW.Find(&workflowList).Error; err != nil {
		return 0, err
	}

	workflows := make([]*v1.Workflow, 0, len(workflowList))
	for _, v := range workflowList {
		workflow, err := v.ToAPI()
		if err != nil {
			return 0, fmt.Errorf("convert workflow(%s) failed: %v", v.Name, err)
		}

		matched := true
		for _, sv := range selectors {
			matched = sv.Match(workflow.Labels)
			if matched {
				break
			}
		}
		if !matched {
			continue
		}
		workflows = append(workflows, workflow)
	}
	if len(workflows) == 0 {
		return 0, errors.New("no workflow matched")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&buildS).Error; err != nil {
			return err
		}

		for k, workflow := range workflows {
			status := &v1.Stage{
				BoxID:     box.ID,
				BuildID:   buildS.ID,
				Number:    uint64(k) + 1,
				Phase:     v1.PhasePending,
				Name:      workflow.Name,
				Limit:     workflow.Spec.Concurrency,
				Worker:    *workflow.Worker(),
				DependsOn: workflow.Spec.DependsOn,
			}
			if !workflow.Spec.When.Match(settings) {
				status.Phase = v1.PhaseSkipped
			}

			var statusS storageV1.Stage
			if err := statusS.FromAPI(status); err != nil {
				return err
			}
			if err := tx.Create(&statusS).Error; err != nil {
				return err
			}

			for sk, sv := range workflow.Spec.Steps {
				step := &v1.Step{
					StageID: statusS.ID,
					Number:  uint64(sk) + 1,
					Phase:   v1.PhasePending,
					Name:    sv.Name,
				}
				if status.Phase == v1.PhaseSkipped {
					step.Phase = v1.PhaseSkipped
				}

				stepS := new(storageV1.Step)
				stepS.FromAPI(step)
				if err := tx.Create(stepS).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return buildS.Number, nil
}

func (s *srv) Cancel(ctx context.Context, namespace, name string, number uint64) error {
	db := database.FromContext(ctx)

	boxS := &storageV1.Box{
		Namespace: namespace,
		Name:      name,
	}
	if err := db.Where(boxS).First(boxS).Error; err != nil {
		return err
	}

	buildS := &storageV1.Build{
		BoxID:  boxS.ID,
		Number: number,
	}
	if err := db.Where(buildS).First(buildS).Error; err != nil {
		return err
	}
	build, err := buildS.ToAPI()
	if err != nil {
		return err
	}
	if build.Phase.IsDone() {
		return errors.New("already done")
	}

	var stages []storageV1.Stage
	if err := db.Where(&storageV1.Stage{
		BoxID:   buildS.BoxID,
		BuildID: buildS.ID,
		Phase:   v1.PhasePending.String(),
	}).Find(&stages).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("find pending stages failed: %v", err)
	}

	stageIDSet := sets.New[uint64]()
	for _, v := range stages {
		stageIDSet.Add(v.ID)
	}
	stageIds := stageIDSet.List()

	if len(stageIds) > 0 {
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&storageV1.Stage{}).
				Where("id in (?)", stageIds).
				Where(&storageV1.Stage{Phase: v1.PhasePending.String()}).
				Update("phase", v1.PhaseCanceled).Error; err != nil {
				return fmt.Errorf("cancel pending stages failed: %v", err)
			}
			if err := tx.Model(&storageV1.Step{}).
				Where("stage_id in (?)", stageIds).
				Where(&storageV1.Step{Phase: v1.PhasePending.String()}).
				Update("phase", v1.PhaseCanceled).Error; err != nil {
				return fmt.Errorf("cancel pending steps failed: %v", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	sched := scheduler.FromContext(ctx)
	return sched.Cancel(ctx, int64(build.ID))
}
