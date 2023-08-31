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
)

type canceller struct {
	sync.Mutex

	subscribers map[chan struct{}]int64
	canceled    map[int64]time.Time
}

func newCanceller() *canceller {
	return &canceller{
		subscribers: make(map[chan struct{}]int64),
		canceled:    make(map[int64]time.Time),
	}
}

func (c *canceller) Cancel(_ context.Context, id int64) error {
	c.Lock()
	c.canceled[id] = time.Now().Add(time.Minute * 5)
	for subscriber, build := range c.subscribers {
		if id == build {
			close(subscriber)
		}
	}
	c.collect()
	c.Unlock()
	return nil
}

func (c *canceller) Canceled(ctx context.Context, id int64) (bool, error) {
	subscriber := make(chan struct{})
	c.Lock()
	c.subscribers[subscriber] = id
	c.Unlock()

	defer func() {
		c.Lock()
		delete(c.subscribers, subscriber)
		c.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(time.Second * 10):
			c.Lock()
			_, ok := c.canceled[id]
			c.Unlock()
			if ok {
				return true, nil
			}
		case <-subscriber:
			return true, nil
		}
	}
}

func (c *canceller) collect() {
	now := time.Now()
	for build, timestamp := range c.canceled {
		if now.After(timestamp) {
			delete(c.canceled, build)
		}
	}
}
