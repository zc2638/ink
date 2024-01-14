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

package clients

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/livelog"
)

type Worker interface {
	V1() WorkerV1
}

type WorkerV1 interface {
	Name() string
	Status(ctx context.Context) error
	Request(ctx context.Context) (*v1.Stage, error)
	Accept(ctx context.Context, stageID uint64) error
	Info(ctx context.Context, stageID uint64) (*v1.Data, error)
	StageBegin(ctx context.Context, stage *v1.Stage) error
	StageEnd(ctx context.Context, stage *v1.Stage) error
	StepBegin(ctx context.Context, step *v1.Step) error
	StepEnd(ctx context.Context, step *v1.Step) error
	LogUpload(ctx context.Context, stepID uint64, lines []*livelog.Line, isAll bool) error
	WatchCancel(ctx context.Context, buildID uint64) error
}

func NewWorker(addr, name string, worker *v1.Worker) (Worker, error) {
	if err := validateURI(addr); err != nil {
		return nil, fmt.Errorf("validate uri failed: %v", err)
	}
	if len(name) == 0 {
		return nil, errors.New("client name must be defined and unique")
	}

	if worker == nil {
		worker = new(v1.Worker)
	}
	if worker.Platform == nil {
		worker.Platform = new(v1.Platform)
	}
	if worker.Platform.OS == "" {
		worker.Platform.OS = runtime.GOOS
	}
	if worker.Platform.Arch == "" {
		worker.Platform.Arch = runtime.GOARCH
	}

	return &client{
		Address: addr,
		name:    name,
		worker:  worker,
	}, nil
}

type client struct {
	Address string
	name    string
	index   int

	worker *v1.Worker
}

func (c *client) V1() WorkerV1 {
	addr := strings.TrimSuffix(c.Address, "/")
	rc := resty.New().SetBaseURL(addr + "/api/client/v1")
	name := c.name + "." + strconv.Itoa(c.index)
	c.index++
	return &clientV1{rc: rc, name: name, worker: c.worker}
}

type clientV1 struct {
	rc     *resty.Client
	name   string
	worker *v1.Worker
}

func (c *clientV1) R(ctx context.Context) *resty.Request {
	return c.rc.R().SetContext(ctx)
}

func (c *clientV1) Name() string {
	return c.name
}

func (c *clientV1) Status(ctx context.Context) error {
	resp, err := c.R(ctx).Post("/status")
	return handleClientError(resp, err)
}

func (c *clientV1) Request(ctx context.Context) (*v1.Stage, error) {
	var result v1.Stage
	req := c.R(ctx).SetBody(c.worker).SetResult(&result)
	resp, err := req.Post("/stage")
	if err := handleClientError(resp, err); err != nil {
		if err == context.DeadlineExceeded {
			return c.Request(ctx)
		}
		return nil, err
	}
	return &result, nil
}

func (c *clientV1) Accept(ctx context.Context, stageID uint64) error {
	req := c.R(ctx).
		SetPathParam("stage", strconv.FormatUint(stageID, 10)).
		SetQueryParam("name", c.name)
	resp, err := req.Post("/stage/{stage}")
	return handleClientError(resp, err)
}

func (c *clientV1) Info(ctx context.Context, stageID uint64) (*v1.Data, error) {
	var result v1.Data
	req := c.R(ctx).
		SetPathParam("stage", strconv.FormatUint(stageID, 10)).
		SetResult(&result)
	resp, err := req.Get("/stage/{stage}")
	if err := handleClientError(resp, err); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *clientV1) StageBegin(ctx context.Context, stage *v1.Stage) error {
	req := c.R(ctx).
		SetBody(stage).
		SetPathParam("stage", strconv.FormatUint(stage.ID, 10))
	resp, err := req.Post("/stage/{stage}/begin")
	return handleClientError(resp, err)
}

func (c *clientV1) StageEnd(ctx context.Context, stage *v1.Stage) error {
	req := c.R(ctx).
		SetBody(stage).
		SetPathParam("stage", strconv.FormatUint(stage.ID, 10))
	resp, err := req.Post("/stage/{stage}/end")
	return handleClientError(resp, err)
}

func (c *clientV1) StepBegin(ctx context.Context, step *v1.Step) error {
	req := c.R(ctx).
		SetBody(step).
		SetPathParam("step", strconv.FormatUint(step.ID, 10))
	resp, err := req.Post("/step/{step}/begin")
	return handleClientError(resp, err)
}

func (c *clientV1) StepEnd(ctx context.Context, step *v1.Step) error {
	req := c.R(ctx).
		SetBody(step).
		SetPathParam("step", strconv.FormatUint(step.ID, 10))
	resp, err := req.Post("/step/{step}/end")
	return handleClientError(resp, err)
}

func (c *clientV1) LogUpload(ctx context.Context, stepID uint64, lines []*livelog.Line, isAll bool) error {
	req := c.R(ctx).
		SetBody(lines).
		SetPathParam("step", strconv.FormatUint(stepID, 10))
	if isAll {
		req = req.SetQueryParam("all", strconv.FormatBool(true))
	}
	resp, err := req.Post("/step/{step}/logs/upload")
	return handleClientError(resp, err)
}

func (c *clientV1) WatchCancel(ctx context.Context, buildID uint64) error {
	req := c.R(ctx).SetPathParam("build", strconv.FormatUint(buildID, 10))
	resp, err := req.Post("/build/{build}/watch")
	if err := handleClientError(resp, err); err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return c.WatchCancel(ctx, buildID)
		}
		return err
	}
	return nil
}
