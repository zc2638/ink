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

package livelog

import (
	"context"
	"errors"
	"sync"
)

const LineMaxBuffer = 3000

type Interface interface {
	List(ctx context.Context, id string) ([]*Line, error)
	Watch(ctx context.Context, id string) (<-chan *Line, <-chan struct{}, error)
	Write(ctx context.Context, id string, line *Line) error
	LineCount(ctx context.Context, id string) int
	Reset(ctx context.Context, id string) error
	Create(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type Config struct {
	File *ConfigFile `json:"file,omitempty"`
}

func New(cfg Config) (Interface, error) {
	if cfg.File != nil {
		return NewFile(*cfg.File)
	}
	return nil, errors.New("no livelog implementation is defined")
}

type Line struct {
	Number  int    `json:"number"`
	Since   int64  `json:"since"`
	Content string `json:"content"`
}

func newSubscriber() *subscriber {
	return &subscriber{
		handler: make(chan *Line, LineMaxBuffer),
		closeCh: make(chan struct{}),
	}
}

type subscriber struct {
	sync.Mutex

	handler chan *Line
	closeCh chan struct{}
	closed  bool
}

func (s *subscriber) publish(line *Line) {
	select {
	case <-s.closeCh:
	case s.handler <- line:
	default:
		// If buffering is exhausted, the message is discarded.
	}
}

func (s *subscriber) waitForClose() <-chan struct{} {
	return s.closeCh
}

func (s *subscriber) close() {
	s.Lock()
	if !s.closed {
		close(s.closeCh)
		s.closed = true
	}
	s.Unlock()
}
