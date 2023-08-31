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

package scheduler

import (
	"context"
	"sync"
	"time"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

type StoreFunc func(ctx context.Context) ([]*v1.StageStatus, error)

type queue struct {
	sync.Mutex

	ready     chan struct{}
	paused    bool
	interval  time.Duration
	storeFunc StoreFunc
	workers   map[*worker]struct{}
	ctx       context.Context
}

// newQueue returns a new Queue backed by the build datastore.
func newQueue(storeFunc StoreFunc) *queue {
	q := &queue{
		storeFunc: storeFunc,
		ready:     make(chan struct{}, 1),
		workers:   map[*worker]struct{}{},
		interval:  time.Minute,
		ctx:       context.Background(),
	}
	go q.start()
	return q
}

func (q *queue) Schedule(ctx context.Context) {
	select {
	case <-ctx.Done():
	case q.ready <- struct{}{}:
	default:
	}
}

func (q *queue) Pause(_ context.Context) error {
	q.Lock()
	q.paused = true
	q.Unlock()
	return nil
}

func (q *queue) Paused(_ context.Context) (bool, error) {
	q.Lock()
	paused := q.paused
	q.Unlock()
	return paused, nil
}

func (q *queue) Resume(ctx context.Context) error {
	q.Lock()
	q.paused = false
	q.Unlock()

	select {
	case <-ctx.Done():
	case q.ready <- struct{}{}:
	default:
	}
	return nil
}

func (q *queue) Request(ctx context.Context, params v1.Worker) (*v1.StageStatus, error) {
	if params.Kind == "" {
		params.Kind = v1.WorkerKindDocker
	}

	w := &worker{
		kind:    params.Kind,
		labels:  params.Labels,
		channel: make(chan *v1.StageStatus),
	}
	if params.Platform != nil {
		w.os = params.Platform.OS
		w.arch = params.Platform.Arch
	}

	q.Lock()
	q.workers[w] = struct{}{}
	q.Unlock()

	select {
	case q.ready <- struct{}{}:
	default:
	}

	select {
	case <-ctx.Done():
		q.Lock()
		delete(q.workers, w)
		q.Unlock()
		return nil, ctx.Err()
	case b := <-w.channel:
		return b, nil
	}
}

func (q *queue) signal(ctx context.Context) error {
	q.Lock()
	count := len(q.workers)
	pause := q.paused
	q.Unlock()
	if pause {
		return nil
	}
	if count == 0 {
		return nil
	}
	if q.storeFunc == nil {
		return nil
	}

	items, err := q.storeFunc(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	q.Lock()
	defer q.Unlock()
	for _, item := range items {
		if item.Phase == v1.PhaseRunning {
			continue
		}

		// if the stage defines concurrency limits, we
		// need to make sure those limits are not exceeded
		// before proceeding.
		if !withinLimits(item, items) {
			continue
		}

		// if the system defines concurrency limits
		// per repository, we need to make sure those limits
		// are not exceeded before proceeding.
		if shouldThrottle(item, items, item.Limit) {
			continue
		}

		for w := range q.workers {
			// the worker must match the resource kind
			if w.kind != item.Worker.Kind {
				continue
			}

			if w.os != "" || w.arch != "" {
				// the worker is platform-specific. check to ensure
				// the queue item matches the worker platform.
				if item.Worker.Platform != nil {
					if w.os != item.Worker.Platform.OS {
						continue
					}
					if w.arch != item.Worker.Platform.Arch {
						continue
					}
				}
			}

			if !checkLabels(item.Worker.Labels, w.labels) {
				continue
			}

			w.channel <- item
			delete(q.workers, w)
			break
		}
	}
	return nil
}

func (q *queue) start() {
	for {
		select {
		case <-q.ctx.Done():
			// TODO q.ctx.Err()
		case <-q.ready:
			_ = q.signal(q.ctx)
		case <-time.After(q.interval):
			_ = q.signal(q.ctx)
		}
	}
}

type worker struct {
	kind    v1.WorkerKind
	os      string
	arch    string
	labels  map[string]string
	channel chan *v1.StageStatus
}

func checkLabels(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

func withinLimits(stage *v1.StageStatus, siblings []*v1.StageStatus) bool {
	if stage.Limit == 0 {
		return true
	}
	count := 0
	for _, sibling := range siblings {
		if sibling.BoxID != stage.BoxID {
			continue
		}
		if sibling.ID == stage.ID {
			continue
		}
		if sibling.Name != stage.Name {
			continue
		}
		if sibling.ID < stage.ID ||
			sibling.Phase == v1.PhaseRunning {
			count++
		}
	}
	return count < stage.Limit
}

func shouldThrottle(stage *v1.StageStatus, siblings []*v1.StageStatus, limit int) bool {
	// if no throttle limit is defined (default) then
	// return false to indicate no throttling is needed.
	if limit == 0 {
		return false
	}
	// if the repository is running, it is too late
	// to skip and we can exit
	if stage.Phase == v1.PhaseRunning {
		return false
	}

	count := 0
	// loop through running stages to count the number of
	// running stages for the parent repository.
	for _, sibling := range siblings {
		// ignore stages from another repository.
		if sibling.BoxID != stage.BoxID {
			continue
		}
		// ignore this stage and stages that were
		// scheduled after this stage.
		if sibling.ID >= stage.ID {
			continue
		}
		count++
	}
	// if the count of running stages exceeds the
	// throttle limit return true.
	return count >= limit
}
