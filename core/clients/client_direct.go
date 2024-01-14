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
	"sync"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
	"github.com/zc2638/ink/pkg/livelog"
)

func NewClientDirect(dataCh chan *v1.Data) WorkerV1 {
	return &clientDirect{
		dataCh: dataCh,
		ds:     make(map[uint64]*v1.Data),
	}
}

type clientDirect struct {
	mux    sync.Mutex
	dataCh chan *v1.Data
	ds     map[uint64]*v1.Data
}

func (c *clientDirect) Name() string {
	return "direct"
}

func (c *clientDirect) Status(_ context.Context) error {
	return nil
}

func (c *clientDirect) Request(ctx context.Context) (*v1.Stage, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-c.dataCh:
		c.mux.Lock()
		defer c.mux.Unlock()

		c.ds[data.Status.ID] = data
		return data.Status, nil
	}
}

func (c *clientDirect) Accept(ctx context.Context, stageID uint64) error {
	_, err := c.Info(ctx, stageID)
	return err
}

func (c *clientDirect) Info(_ context.Context, stageID uint64) (*v1.Data, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	data, ok := c.ds[stageID]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}

func (c *clientDirect) StageBegin(_ context.Context, _ *v1.Stage) error {
	return nil
}

func (c *clientDirect) StageEnd(_ context.Context, _ *v1.Stage) error {
	return nil
}

func (c *clientDirect) StepBegin(_ context.Context, step *v1.Step) error {
	fmt.Printf("\x1b[1m%s\x1b[0m\n", "[STEP] "+step.Name)
	return nil
}

func (c *clientDirect) StepEnd(_ context.Context, _ *v1.Step) error {
	return nil
}

func (c *clientDirect) LogUpload(_ context.Context, _ uint64, lines []*livelog.Line, isAll bool) error {
	if isAll {
		return nil
	}

	for _, line := range lines {
		fmt.Print(line.Content)
	}
	return nil
}

func (c *clientDirect) WatchCancel(ctx context.Context, _ uint64) error {
	<-ctx.Done()
	return ctx.Err()
}
