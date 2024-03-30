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

package worker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/99nil/gopkg/sets"
	"github.com/zc2638/wslog"
	"golang.org/x/sync/errgroup"

	"github.com/zc2638/ink/core/clients"
	"github.com/zc2638/ink/core/constant"
	"github.com/zc2638/ink/core/worker/runtime"
	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/livelog"
)

type Hook interface {
	Begin(ctx context.Context, spec *Workflow) error
	End(ctx context.Context, spec *Workflow) error
	Step(ctx context.Context, spec *Workflow, step *Step, writer io.Writer) (*State, error)
}

func NewMust(client clients.Worker, hook Hook, log *wslog.Logger, count int) *Worker {
	w, err := New(client, hook, log, count)
	if err != nil {
		panic(err)
	}
	return w
}

func New(client clients.Worker, hook Hook, log *wslog.Logger, count int) (*Worker, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	if hook == nil {
		return nil, errors.New("hook is nil")
	}
	if log == nil {
		log = wslog.Default()
	}

	w := &Worker{
		log:    log,
		client: client,
		hook:   hook,
		count:  1,
	}
	if count > 0 {
		w.count = count
	}
	return w, nil
}

type Worker struct {
	log    *wslog.Logger
	client clients.Worker
	hook   Hook
	count  int
}

func (w *Worker) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < w.count; i++ {
		clientV1 := w.client.V1()
		log := w.log.With("client_name", clientV1.Name())
		wCtx := wslog.WithContext(ctx, log)

		eg.Go(func() error {
			var waitTimes int
			for {
				select {
				case <-wCtx.Done():
					return wCtx.Err()
				default:
				}

				if err := Run(wCtx, clientV1, w.hook); err != nil {
					log.Error("Run worker failed",
						"error", err,
						"wait", waitTimes,
					)

					waitSec := math.Pow(2, float64(waitTimes))
					if waitSec > 60 {
						waitSec = constant.DefaultWaitTime
					}
					waitTimes++

					select {
					case <-time.After(time.Second * time.Duration(waitSec)):
					case <-wCtx.Done():
						return wCtx.Err()
					}
					continue
				}
				waitTimes = 0
				log.Debug("Run worker success")
			}
		})
	}
	return eg.Wait()
}

func Run(ctx context.Context, client clients.WorkerV1, hook Hook) error {
	log := wslog.FromContext(ctx)

	log.Debug("Request stage")
	stage, err := client.Request(ctx)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	log = log.With(
		"stage_name", stage.Name,
		"stage_id", stage.ID,
	)
	log.Debug("Request stage success")

	if err := client.Accept(ctx, stage.ID); err != nil {
		return fmt.Errorf("accept failed: %v", err)
	}
	log.Debug("Accept stage success")

	data, err := client.Info(ctx, stage.ID)
	if err != nil {
		return fmt.Errorf("get data failed: %v", err)
	}
	workflow := data.Workflow
	status := data.Status
	if workflow == nil || status == nil {
		return errors.New("stage data not found")
	}

	log = log.With("namespace", workflow.GetNamespace())
	log.Debug("Get data success", "build_id", status.BuildID)

	var settings map[string]string
	if data.Build != nil {
		settings = data.Build.CompleteSettings(data.Box)
	}

	ctx = wslog.WithContext(ctx, log)
	eg, runCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := client.WatchCancel(runCtx, status.BuildID)
		if err == nil {
			return v1.ErrCanceled
		}
		return err
	})
	eg.Go(func() error {
		if err := execute(runCtx, client, hook, workflow, status, data.Secrets, settings); err != nil {
			return err
		}
		return context.Canceled
	})
	err = eg.Wait()
	if err == context.Canceled {
		return nil
	}
	if errors.Is(err, v1.ErrCanceled) {
		return cancel(ctx, client, status)
	}
	return err
}

func execute(
	ctx context.Context,
	client clients.WorkerV1,
	hook Hook,
	workflow *v1.Workflow,
	status *v1.Stage,
	secrets []*v1.Secret,
	settings map[string]string,
) error {
	log := wslog.FromContext(ctx)
	log.Debug("Execute stage begin request")

	spec, err := Convert(workflow, status, secrets)
	if err != nil {
		return fmt.Errorf("convert to worker stage failed: %v", err)
	}

	status.Started = time.Now().Unix()
	if !workflow.Spec.When.Match(settings) {
		status.Phase = v1.PhaseSkipped
		for _, step := range status.Steps {
			step.Phase = v1.PhaseSkipped
			step.Started = status.Started
		}
		if err := client.StageEnd(ctx, status); err != nil {
			return fmt.Errorf("skip stage failed: %v", err)
		}
		return nil
	}

	secretValueSet := sets.New[string]()
	for _, secret := range secrets {
		if err := secret.Decrypt(); err != nil {
			status.Phase = v1.PhaseFailed
			status.Error = err.Error()
			for _, step := range status.Steps {
				step.Phase = v1.PhaseSkipped
				step.Started = status.Started
			}
			if err := client.StageEnd(ctx, status); err != nil {
				return fmt.Errorf("stage end failed: %v", err)
			}
			return nil
		}
		for _, v := range secret.Data {
			secretValueSet.Add(v)
		}
	}
	secretValueList := secretValueSet.List()

	var (
		failed   bool
		canceled bool
	)
	status.Phase = v1.PhaseRunning
	if err := client.StageBegin(ctx, status); err != nil {
		return fmt.Errorf("stage begin request failed: %v", err)
	}

	log.Debug("Execute stage begin hook")
	if err := hook.Begin(ctx, spec); err != nil {
		failed = true
		status.Error = err.Error()
		log.Error("Execute stage begin hook failed", "error", err)
	}

	for _, step := range status.Steps {
		stepSpec := spec.GetStep(step.Name)
		if stepSpec != nil {
			stepSpec.Env = stepSpec.CombineEnv(settings)
		}

		stepLog := log.With(
			"step_id", step.ID,
			"step_name", step.Name,
		)

		step.Started = time.Now().Unix()
		if stepSpec == nil || failed {
			step.Phase = v1.PhaseSkipped
			step.Stopped = step.Started

			stepLog.Debug("Execute step end request by skipped")
			if err := client.StepEnd(ctx, step); err != nil {
				return fmt.Errorf("step(%s) end request failed: %v", step.Name, err)
			}
			continue
		}

		if canceled {
			step.Phase = v1.PhaseCanceled

			stepLog.Debug("Execute step end request by canceled")
			if err := client.StepEnd(ctx, step); err != nil {
				return fmt.Errorf("step(%s) end request failed: %v", step.Name, err)
			}
			continue
		}
		step.Phase = v1.PhaseRunning

		stepLog.Debug("Execute step begin request")
		if err := client.StepBegin(ctx, step); err != nil {
			return fmt.Errorf("step(%s) begin request failed: %v", step.Name, err)
		}

		logHandle := func(lines []*livelog.Line, isAll bool) {
			if len(lines) == 0 {
				return
			}
			if err := client.LogUpload(ctx, step.ID, lines, isAll); err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					stepLog.Debug("Upload log canceled")
					return
				}
				stepLog.Error("Upload log failed", "error", err)
			}
		}
		wc := livelog.NewWriter(logHandle)
		wc = runtime.NewMaskReplacer(wc, secretValueList)

		stepLog.Debug("Execute step hook")
		state, err := hook.Step(ctx, spec, stepSpec, wc)
		_ = wc.Close()

		step.Phase = v1.PhaseSucceeded
		step.Stopped = time.Now().Unix()
		if errors.Is(err, context.Canceled) {
			step.Phase = v1.PhaseCanceled
			canceled = true
			stepLog.Debug("Execute step hook cancel")
		} else if err != nil {
			step.Phase = v1.PhaseFailed
			step.Error = err.Error()
			failed = true
			stepLog.Error("Execute step hook failed")
		}

		if state != nil {
			if state.OOMKilled {
				stepLog.Debug("received oom kill.")
				state.ExitCode = 137
			} else {
				stepLog.Debugf("received exit code %d", state.ExitCode)
			}
			// if the exit code is 78, the system will skip all
			// subsequent pending steps in the pipeline.
			if state.ExitCode == 78 {
				stepLog.Debug("received exit code 78. early exit.")
				step.Phase = v1.PhaseSkipped
				failed = true
			} else if state.ExitCode > 0 {
				step.Phase = v1.PhaseFailed
				failed = true
			}
		}

		stepLog.Debug("Execute step end request")
		if err := client.StepEnd(ctx, step); err != nil {
			return fmt.Errorf("step(%s) end request failed: %v", step.Name, err)
		}
	}

	log.Debug("Execute stage end hook")
	if err := hook.End(ctx, spec); err != nil {
		log.Error("Execute stage end hook failed", "error", err)
	}

	status.Stopped = time.Now().Unix()
	status.Phase = v1.PhaseSucceeded
	if failed {
		status.Phase = v1.PhaseFailed
	}

	log.Debug("Execute stage end request")
	if err := client.StageEnd(ctx, status); err != nil {
		return fmt.Errorf("stage end request failed: %v", err)
	}
	return nil
}

func cancel(ctx context.Context, client clients.WorkerV1, status *v1.Stage) error {
	if status.Phase.IsDone() {
		return nil
	}

	for _, step := range status.Steps {
		if step.Phase.IsDone() {
			continue
		}
		step.Phase = v1.PhaseCanceled
		step.Stopped = time.Now().Unix()
		if step.Started == 0 {
			step.Started = step.Stopped
		}
		if err := client.StepEnd(ctx, step); err != nil {
			return err
		}
	}

	status.Phase = v1.PhaseCanceled
	status.Stopped = time.Now().Unix()
	if status.Started == 0 {
		status.Started = status.Stopped
	}
	if err := client.StageEnd(ctx, status); err != nil {
		return fmt.Errorf("stage end request failed: %v", err)
	}
	return nil
}
