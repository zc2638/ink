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

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/99nil/gopkg/ctr"
	"github.com/99nil/gopkg/sets"
	"gorm.io/gorm"

	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/handler/wrapper"
	"github.com/zc2638/ink/core/scheduler"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	storageV1 "github.com/zc2638/ink/pkg/api/storage/v1"
	"github.com/zc2638/ink/pkg/database"
	"github.com/zc2638/ink/pkg/livelog"
)

// handleStatus returns a `http.HandlerFunc`
// that makes a `http.Request` to report client status.
func handleStatus(w http.ResponseWriter, _ *http.Request) {
	ctr.OK(w, "ok")
}

// handleRequest returns a `http.HandlerFunc`
// that processes a `http.Request` to request a stage from the queue for execution.
func handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), constant.DefaultHTTPTimeout)
	defer cancel()

	var worker v1.Worker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		ctr.BadRequest(w, err)
		return
	}

	sched := scheduler.FromRequest(r)
	stage, err := sched.Request(ctx, worker)
	if err != nil {
		ctr.InternalError(w, err)
		return
	}
	ctr.OK(w, stage)
}

// handleAccept returns a `http.HandlerFunc`
// that processes a `http.Request` to accept ownership of the stage.
func handleAccept() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stageID, _ := strconv.ParseUint(
			wrapper.URLParam(r, "stage"), 10, 64)
		query := r.URL.Query()
		workerName := query.Get("name")
		db := database.FromRequest(r)

		stageS := new(storageV1.Stage)
		stageS.SetID(stageID)
		if err := db.Where(stageS).First(stageS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		if stageS.WorkerName != "" && stageS.WorkerName != workerName {
			wrapper.BadRequest(w, "stage already assigned. abort")
			return
		}

		stageS.WorkerName = workerName
		stageS.Phase = v1.PhasePending.String()
		if err := db.Model(stageS).Updates(stageS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

// handleInfo returns a `http.HandlerFunc`
// that processes a `http.Request` to get the build details.
func handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stageID, _ := strconv.ParseUint(
			wrapper.URLParam(r, "stage"), 10, 64)
		db := database.FromRequest(r)

		statusS := new(storageV1.Stage)
		statusS.SetID(stageID)
		if err := db.Where(statusS).First(statusS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		status, err := statusS.ToAPI()
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		var stepList []storageV1.Step
		if err := db.Where(&storageV1.Step{StageID: status.ID}).Find(&stepList).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		for _, v := range stepList {
			status.Steps = append(status.Steps, v.ToAPI())
		}

		buildS := new(storageV1.Build)
		buildS.SetID(status.BuildID)
		if err := db.Where(buildS).First(buildS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		build, err := buildS.ToAPI()
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		var stageList []storageV1.Stage
		if err := db.Where(&storageV1.Stage{BuildID: build.ID}).Find(&stageList).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		for _, v := range stageList {
			item, err := v.ToAPI()
			if err != nil {
				wrapper.InternalError(w, err)
				return
			}
			build.Stages = append(build.Stages, item)
		}

		boxS := new(storageV1.Box)
		boxS.SetID(build.BoxID)
		if err := db.Where(boxS).First(boxS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		box, err := boxS.ToAPI()
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		stageS := &storageV1.Workflow{Namespace: box.GetNamespace(), Name: status.Name}
		if err := db.Where(stageS).First(stageS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stage, err := stageS.ToAPI()
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		data := &v1.Data{
			Box:      box,
			Build:    build,
			Workflow: stage,
			Status:   status,
		}

		// get secrets
		var secretList []storageV1.Secret
		secretNames, selectors := box.GetSelectors(v1.KindSecret, build.Settings)
		if len(secretNames) > 0 {
			secretDB := db.Where(&storageV1.Secret{Namespace: box.GetNamespace()})
			if !slices.Contains(secretNames, "") {
				secretDB = secretDB.Where("name in (?)", secretNames)
			}
			if err := secretDB.Find(&secretList).Error; err != nil {
				wrapper.InternalError(w, err)
				return
			}
		}
		for _, v := range secretList {
			secret, err := v.ToAPI()
			if err != nil {
				wrapper.InternalError(w, err)
				return
			}

			matched := true
			for _, sv := range selectors {
				matched = sv.Match(secret.Labels)
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
			data.Secrets = append(data.Secrets, secret)
		}
		ctr.OK(w, data)
	}
}

// handleStageBegin returns a `http.HandlerFunc`
// that processes a `http.Request` to update the stage status.
func handleStageBegin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stage := new(v1.Stage)
		err := json.NewDecoder(r.Body).Decode(&stage)
		if err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		if stage.Phase != v1.PhasePending && stage.Phase != v1.PhaseRunning {
			wrapper.BadRequest(w, "the stage has already begun")
			return
		}

		db := database.FromRequest(r)

		buildS := new(storageV1.Build)
		buildS.SetID(stage.BuildID)
		if err := db.Where(buildS).First(buildS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		if len(stage.Error) > 500 {
			stage.Error = stage.Error[:500]
		}
		stageS := new(storageV1.Stage)
		if err := stageS.FromAPI(stage); err != nil {
			wrapper.InternalError(w, err)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(stageS).Updates(stageS).Error; err != nil {
				return err
			}
			if buildS.Phase != v1.PhasePending.String() {
				return nil
			}
			buildS.Started = time.Now().Unix()
			buildS.Phase = v1.PhaseRunning.String()
			buildWhere := new(storageV1.Build)
			buildWhere.SetID(buildS.ID)
			return tx.Model(buildWhere).Where(buildWhere).Updates(buildS).Error
		})
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.Success(w)
	}
}

// handleStageEnd returns a `http.HandlerFunc`
// that processes a `http.Request` to update the stage status.
func handleStageEnd() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stage := new(v1.Stage)
		err := json.NewDecoder(r.Body).Decode(&stage)
		if err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		if stage.Phase == v1.PhasePending {
			wrapper.BadRequest(w, "the stage has not yet begun")
			return
		}

		ctx := r.Context()
		db := database.FromContext(ctx)
		ll := livelog.FromContext(ctx)
		sched := scheduler.FromContext(ctx)

		buildS := new(storageV1.Build)
		buildS.SetID(stage.BuildID)
		if err := db.Where(buildS).First(buildS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}

		if len(stage.Error) > 500 {
			stage.Error = stage.Error[:500]
		}
		stageS := new(storageV1.Stage)
		if err := stageS.FromAPI(stage); err != nil {
			wrapper.InternalError(w, err)
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			stageWhere := new(storageV1.Stage)
			stageWhere.SetID(stageS.ID)
			if err := tx.Model(stageWhere).Where(stageWhere).Updates(stageS).Error; err != nil {
				return err
			}

			for _, step := range stage.Steps {
				if len(step.Error) > 500 {
					step.Error = step.Error[:500]
				}

				stepS := new(storageV1.Step)
				stepS.FromAPI(step)
				stepWhere := new(storageV1.Step)
				stepWhere.SetID(stepS.ID)
				if err := tx.Model(stepWhere).Where(stepWhere).Updates(stepS).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}

		for _, step := range stage.Steps {
			// TODO need to log
			_ = ll.Delete(ctx, strconv.FormatUint(step.ID, 10))
		}

		var stageList []storageV1.Stage
		if err := db.Where(&storageV1.Stage{BuildID: buildS.ID}).Find(&stageList).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		stages := make([]*v1.Stage, 0, len(stageList))
		for _, v := range stageList {
			item, err := v.ToAPI()
			if err != nil {
				wrapper.InternalError(w, err)
				return
			}
			stages = append(stages, item)
		}

		if err := cancelDownstream(db, stages); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		if err := scheduleDownstream(ctx, sched, db, stages); err != nil {
			wrapper.InternalError(w, err)
			return
		}

		isBuildComplete := true
		for _, sv := range stages {
			if sv.Phase == v1.PhaseUnknown ||
				sv.Phase == v1.PhaseWaiting ||
				sv.Phase == v1.PhasePending ||
				sv.Phase == v1.PhaseRunning {
				isBuildComplete = false
				break
			}
		}
		if isBuildComplete {
			buildS.Phase = v1.PhaseSucceeded.String()
			buildS.Stopped = time.Now().Unix()
			for _, sv := range stages {
				if sv.Phase == v1.PhaseFailed || sv.Phase == v1.PhaseCanceled {
					buildS.Phase = sv.Phase.String()
					break
				}
			}

			if buildS.Started == 0 {
				buildS.Started = buildS.Stopped
			}
			buildWhere := new(storageV1.Build)
			buildWhere.SetID(buildS.ID)
			if err := db.Model(buildWhere).Where(buildWhere).Updates(buildS).Error; err != nil {
				wrapper.InternalError(w, err)
				return
			}
		}

		ctr.Success(w)
	}
}

// handleStepBegin returns a `http.HandlerFunc`
// that processes a `http.Request` to update the step status.
func handleStepBegin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step := new(v1.Step)
		err := json.NewDecoder(r.Body).Decode(&step)
		if err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		if step.Phase != v1.PhasePending && step.Phase != v1.PhaseRunning {
			wrapper.BadRequest(w, "the step has already begun")
			return
		}

		ctx := r.Context()
		ll := livelog.FromContext(ctx)
		db := database.FromContext(ctx)

		if err := ll.Create(ctx, strconv.FormatUint(step.ID, 10)); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		if len(step.Error) > 500 {
			step.Error = step.Error[:500]
		}

		stepS := new(storageV1.Step)
		stepS.FromAPI(step)
		stepWhere := new(storageV1.Step)
		stepWhere.SetID(step.ID)
		if err := db.Model(stepWhere).Where(stepWhere).Updates(stepS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, step)
	}
}

// handleStepEnd returns a `http.HandlerFunc`
// that processes a `http.Request` to update the step status.
func handleStepEnd() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step := new(v1.Step)
		err := json.NewDecoder(r.Body).Decode(&step)
		if err != nil {
			wrapper.BadRequest(w, err)
			return
		}
		if step.Phase == v1.PhasePending {
			wrapper.BadRequest(w, "the step has not yet begun")
			return
		}

		ctx := r.Context()
		ll := livelog.FromContext(ctx)
		db := database.FromContext(ctx)

		if len(step.Error) > 500 {
			step.Error = step.Error[:500]
		}
		stepS := new(storageV1.Step)
		stepS.FromAPI(step)
		stepWhere := new(storageV1.Step)
		stepWhere.SetID(step.ID)
		if err := db.Model(stepWhere).Where(stepWhere).Updates(stepS).Error; err != nil {
			wrapper.InternalError(w, err)
			return
		}

		lines, err := ll.List(ctx, strconv.FormatUint(step.ID, 10))
		if err != nil {
			wrapper.InternalError(w, err)
			return
		}
		if len(lines) > 0 {
			logBytes, err := json.Marshal(lines)
			if err != nil {
				wrapper.InternalError(w, err)
				return
			}

			logS := new(storageV1.Log)
			logS.SetID(step.ID)
			logS.Data = logBytes
			if err := db.Create(logS).Error; err != nil {
				wrapper.InternalError(w, err)
				return
			}
		}

		if err := ll.Delete(ctx, strconv.FormatUint(step.ID, 10)); err != nil {
			wrapper.InternalError(w, err)
			return
		}
		ctr.OK(w, step)
	}
}

// handleLogUpload returns a `http.HandlerFunc`
// that accepts a `http.Request` to submit a stream of logs to the server.
func handleLogUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stepID := wrapper.URLParam(r, "step")
		isAll, _ := strconv.ParseBool(r.URL.Query().Get("all"))

		var lines []*livelog.Line
		if err := json.NewDecoder(r.Body).Decode(&lines); err != nil {
			ctr.BadRequest(w, err)
			return
		}
		if len(lines) == 0 {
			ctr.BadRequest(w, "empty log line")
			return
		}

		ctx := r.Context()
		ll := livelog.FromContext(ctx)

		var opts []any
		if isAll {
			lineCount := ll.LineCount(ctx, stepID)
			if lineCount == len(lines) {
				ctr.Success(w)
				return
			}
			if err := ll.Reset(ctx, stepID); err != nil {
				ctr.InternalError(w, err)
				return
			}
			opts = []any{livelog.PublishOption(false)}
		}
		for _, line := range lines {
			if err := ll.Write(ctx, stepID, line, opts...); err != nil {
				ctr.InternalError(w, err)
				return
			}
		}
		ctr.Success(w)
	}
}

// handleWatch returns a `http.HandlerFunc`
// that accepts a blocking `http.Request` that watches a build for cancellation.
func handleWatchCancel(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), constant.DefaultHTTPTimeout)
	defer cancel()

	buildID, _ := strconv.ParseUint(
		wrapper.URLParam(r, "build"), 10, 64)

	sched := scheduler.FromRequest(r)
	_, err := sched.Canceled(ctx, int64(buildID))

	// Expect a context cancel error here which
	// indicates a polling timeout. The subscribing
	// client should look for the context cancel error
	// and resume polling.
	if err != nil {
		ctr.InternalError(w, err)
		return
	}
	ctr.Success(w)
}

func cancelDownstream(db *gorm.DB, stages []*v1.Stage) error {
	failed := false
	for _, s := range stages {
		if s.Phase.IsFailed() {
			failed = true
		}
	}

	var errs []error
	for _, s := range stages {
		if s.Phase != v1.PhaseWaiting {
			continue
		}

		var skip bool
		if failed && !s.Phase.IsFailed() {
			skip = true
		}
		if !failed && !s.Phase.IsSucceeded() {
			skip = true
		}
		if !skip {
			continue
		}

		if !areDepsComplete(s, stages) {
			continue
		}

		s.Phase = v1.PhaseSkipped
		s.Started = time.Now().Unix()
		s.Stopped = time.Now().Unix()

		stageS := new(storageV1.Stage)
		if err := stageS.FromAPI(s); err != nil {
			errs = append(errs, err)
			continue
		}
		stageWhere := new(storageV1.Stage)
		stageWhere.SetID(s.ID)
		if err := db.Model(stageWhere).Where(stageWhere).Updates(stageS).Error; err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func areDepsComplete(stage *v1.Stage, stages []*v1.Stage) bool {
	deps := sets.NewString(stage.DependsOn...)
	for _, sv := range stages {
		if !deps.Has(sv.Name) {
			continue
		}
		if !sv.Phase.IsDone() {
			return false
		}
	}
	return true
}

func scheduleDownstream(ctx context.Context, sched scheduler.Interface, db *gorm.DB, stages []*v1.Stage) error {
	var errs []error
	for _, sv := range stages {
		if sv.Phase != v1.PhaseWaiting {
			continue
		}
		if len(sv.DependsOn) == 0 {
			continue
		}
		if !areDepsComplete(sv, stages) {
			continue
		}

		sv.Phase = v1.PhasePending

		stageS := new(storageV1.Stage)
		if err := stageS.FromAPI(sv); err != nil {
			errs = append(errs, err)
			continue
		}
		stageWhere := new(storageV1.Stage)
		stageWhere.SetID(sv.ID)
		if err := db.Model(stageWhere).Where(stageWhere).Updates(stageS).Error; err != nil {
			errs = append(errs, err)
			continue
		}
		sched.Schedule(ctx)
	}
	return errors.Join(errs...)
}
