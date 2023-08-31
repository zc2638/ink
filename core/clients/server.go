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
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/livelog"
	"github.com/zc2638/ink/pkg/sse"
)

type Server interface {
	V1() ServerV1
}

type ServerV1 interface {
	StageList(ctx context.Context, page v1.Pagination) ([]*v1.Stage, *v1.Pagination, error)
	StageInfo(ctx context.Context, namespace, name string) (*v1.Stage, error)
	StageCreate(ctx context.Context, data *v1.Stage) error
	StageUpdate(ctx context.Context, data *v1.Stage) error
	StageDelete(ctx context.Context, namespace, name string) error

	BoxList(ctx context.Context, page v1.Pagination) ([]*v1.Box, *v1.Pagination, error)
	BoxInfo(ctx context.Context, namespace, name string) (*v1.Box, error)
	BoxCreate(ctx context.Context, data *v1.Box) error
	BoxUpdate(ctx context.Context, data *v1.Box) error
	BoxDelete(ctx context.Context, namespace, name string) error

	BuildList(ctx context.Context, namespace, name string, page v1.Pagination) ([]*v1.Build, *v1.Pagination, error)
	BuildInfo(ctx context.Context, namespace, name string, number uint64) (*v1.Build, error)
	BuildCreate(ctx context.Context, namespace, name string) (uint64, error)
	BuildCancel(ctx context.Context, namespace, name string, number uint64) error

	LogInfo(ctx context.Context, namespace, name string, number, stage, step uint64) ([]*livelog.Line, error)
	LogWatch(ctx context.Context, namespace, name string, number, stage, step uint64) (<-chan *livelog.Line, <-chan error, error)
}

func NewServer(addr string) (Server, error) {
	if err := validateURI(addr); err != nil {
		return nil, err
	}
	return &server{Address: addr}, nil
}

type server struct {
	Address string
}

func (s *server) V1() ServerV1 {
	addr := strings.TrimSuffix(s.Address, "/")
	rc := resty.New().SetBaseURL(addr + "/api/core/v1")
	return &serverV1{rc: rc}
}

type serverV1 struct {
	rc *resty.Client
}

func (c *serverV1) R(ctx context.Context) *resty.Request {
	return c.rc.R().SetContext(ctx)
}

func (c *serverV1) StageList(ctx context.Context, page v1.Pagination) ([]*v1.Stage, *v1.Pagination, error) {
	type resultT struct {
		v1.Pagination
		Items []*v1.Stage `json:"items"`
	}

	var result resultT
	vs := page.ToValues()
	req := c.R(ctx).SetQueryParamsFromValues(vs).SetResult(&result)
	resp, err := req.Get("/stage")
	if err := handleClientError(resp, err); err != nil {
		return nil, nil, err
	}
	return result.Items, &result.Pagination, nil
}

func (c *serverV1) StageInfo(ctx context.Context, namespace, name string) (*v1.Stage, error) {
	var result v1.Stage
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetResult(&result)
	resp, err := req.Get("/stage/{namespace}/{name}")
	if err := handleClientError(resp, err); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *serverV1) StageCreate(ctx context.Context, data *v1.Stage) error {
	req := c.R(ctx).SetBody(data)
	resp, err := req.Post("/stage")
	return handleClientError(resp, err)
}

func (c *serverV1) StageUpdate(ctx context.Context, data *v1.Stage) error {
	req := c.R(ctx).
		SetBody(data).
		SetPathParam("namespace", data.GetNamespace()).
		SetPathParam("name", data.GetName())
	resp, err := req.Put("/stage/{namespace}/{name}")
	return handleClientError(resp, err)
}

func (c *serverV1) StageDelete(ctx context.Context, namespace, name string) error {
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name)
	resp, err := req.Delete("/stage/{namespace}/{name}")
	return handleClientError(resp, err)
}

func (c *serverV1) BoxList(ctx context.Context, page v1.Pagination) ([]*v1.Box, *v1.Pagination, error) {
	type resultT struct {
		v1.Pagination
		Items []*v1.Box `json:"items"`
	}

	var result resultT
	vs := page.ToValues()
	req := c.R(ctx).SetQueryParamsFromValues(vs).SetResult(&result)
	resp, err := req.Get("/box")
	if err := handleClientError(resp, err); err != nil {
		return nil, nil, err
	}
	return result.Items, &result.Pagination, nil
}

func (c *serverV1) BoxInfo(ctx context.Context, namespace, name string) (*v1.Box, error) {
	var result v1.Box
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetResult(&result)
	resp, err := req.Get("/box/{namespace}/{name}")
	if err := handleClientError(resp, err); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *serverV1) BoxCreate(ctx context.Context, data *v1.Box) error {
	req := c.R(ctx).SetBody(data)
	resp, err := req.Post("/box")
	return handleClientError(resp, err)
}

func (c *serverV1) BoxUpdate(ctx context.Context, data *v1.Box) error {
	req := c.R(ctx).
		SetBody(data).
		SetPathParam("namespace", data.GetNamespace()).
		SetPathParam("name", data.GetName())
	resp, err := req.Put("/box/{namespace}/{name}")
	return handleClientError(resp, err)
}

func (c *serverV1) BoxDelete(ctx context.Context, namespace, name string) error {
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name)
	resp, err := req.Delete("/box/{namespace}/{name}")
	return handleClientError(resp, err)
}

func (c *serverV1) BuildList(ctx context.Context, namespace, name string, page v1.Pagination) ([]*v1.Build, *v1.Pagination, error) {
	type resultT struct {
		v1.Pagination
		Items []*v1.Build `json:"items"`
	}

	var result resultT
	vs := page.ToValues()
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetQueryParamsFromValues(vs).
		SetResult(&result)
	resp, err := req.Get("/box/{namespace}/{name}/build")
	if err := handleClientError(resp, err); err != nil {
		return nil, nil, err
	}
	return result.Items, &result.Pagination, nil
}

func (c *serverV1) BuildInfo(ctx context.Context, namespace, name string, number uint64) (*v1.Build, error) {
	var result v1.Build
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetPathParam("number", strconv.FormatUint(number, 10)).
		SetResult(&result)
	resp, err := req.Get("/box/{namespace}/{name}/build/{number}")
	if err := handleClientError(resp, err); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *serverV1) BuildCreate(ctx context.Context, namespace, name string) (uint64, error) {
	var result uint64
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetResult(&result)
	resp, err := req.Post("/box/{namespace}/{name}/build")
	if err := handleClientError(resp, err); err != nil {
		return 0, err
	}
	return result, nil
}

func (c *serverV1) BuildCancel(ctx context.Context, namespace, name string, number uint64) error {
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetPathParam("number", strconv.FormatUint(number, 10))
	resp, err := req.Post("/box/{namespace}/{name}/build/{number}/cancel")
	return handleClientError(resp, err)
}

func (c *serverV1) LogInfo(ctx context.Context, namespace, name string, number, stage, step uint64) ([]*livelog.Line, error) {
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetPathParam("number", strconv.FormatUint(number, 10)).
		SetPathParam("stage", strconv.FormatUint(stage, 10)).
		SetPathParam("step", strconv.FormatUint(step, 10))
	resp, err := req.Get("/box/{namespace}/{name}/build/{number}/logs/{stage}/{step}")
	if err := handleClientError(resp, err); err != nil {
		return nil, err
	}
	var result []*livelog.Line
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *serverV1) LogWatch(ctx context.Context, namespace, name string, number, stage, step uint64) (<-chan *livelog.Line, <-chan error, error) {
	req := c.R(ctx).
		SetPathParam("namespace", namespace).
		SetPathParam("name", name).
		SetPathParam("number", strconv.FormatUint(number, 10)).
		SetPathParam("stage", strconv.FormatUint(stage, 10)).
		SetPathParam("step", strconv.FormatUint(step, 10)).
		SetDoNotParseResponse(true)
	resp, err := req.Post("/box/{namespace}/{name}/build/{number}/logs/{stage}/{step}")
	if err := handleClientError(resp, err); err != nil {
		return nil, nil, err
	}

	logCh := make(chan *livelog.Line)
	errCh := make(chan error)
	go func() {
		sseParser := sse.NewParser(resp.RawBody())
		err = sseParser.ReadEventLoop(func(message *sse.Message, err error) error {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
			}
			if errors.Is(err, io.EOF) {
				return nil
			}
			if err != nil {
				return err
			}

			if message.Event == "error" {
				if strings.ToLower(message.Data) == "eof" {
					return io.EOF
				}
				return errors.New(message.Data)
			}
			if message.Event != "data" {
				return nil
			}

			var line livelog.Line
			if err := json.Unmarshal([]byte(message.Data), &line); err != nil {
				// TODO log
				return nil
			}
			logCh <- &line
			return nil
		})
		errCh <- err
		close(errCh)
	}()
	return logCh, errCh, nil
}
