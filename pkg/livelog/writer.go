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
	"io"
	"strings"
	"time"
)

func NewWriter(handler func(lines []*Line)) io.WriteCloser {
	w := &writer{
		handler: handler,
		now:     time.Now(),
		lineCh:  make(chan *Line, 1024),
		closeCh: make(chan struct{}),
		readyCh: make(chan struct{}),
	}
	go w.process()
	return w
}

type writer struct {
	handler func([]*Line)

	num     int
	now     time.Time
	lines   []*Line
	lineCh  chan *Line
	closeCh chan struct{}
	readyCh chan struct{}
}

func (w *writer) Write(p []byte) (n int, err error) {
	for _, part := range split(p) {
		line := &Line{
			Number:  w.num,
			Since:   int64(time.Since(w.now).Seconds()),
			Content: part,
		}
		w.num++

		select {
		case w.lineCh <- line:
		case <-w.closeCh:
			break
		}
	}
	return len(p), nil
}

func (w *writer) Close() error {
	close(w.closeCh)
	if w.handler != nil && len(w.lines) > 0 {
		w.handler(w.lines)
	}
	return nil
}

func (w *writer) process() {
	for {
		select {
		case <-w.closeCh:
			return
		case line := <-w.lineCh:
			w.lines = append(w.lines, line)
		case <-w.readyCh:
			<-time.After(time.Second)

			if len(w.lines) > 0 {
				if w.handler != nil {
					w.handler(w.lines)
				}
				w.lines = make([]*Line, 0)
			}
		}
	}
}

func split(p []byte) []string {
	s := string(p)
	v := []string{s}
	if strings.Contains(strings.TrimSuffix(s, "\n"), "\n") {
		v = strings.SplitAfter(s, "\n")
	}
	return v
}
