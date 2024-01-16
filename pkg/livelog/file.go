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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type ConfigFile struct {
	Dir string `json:"dir"`
}

type fileItem struct {
	sync.Mutex
	file  *os.File
	count int
}

func (fi *fileItem) Close() error {
	return fi.file.Close()
}

func (fi *fileItem) Read(ctx context.Context, start, size int) ([]*Line, error) {
	if size < 1 {
		size = 100
	}
	end := size
	if start > 0 {
		end += start
	}

	fi.Lock()
	defer fi.Unlock()

	if _, err := fi.file.Seek(0, 0); err != nil {
		return nil, err
	}

	lines := make([]*Line, 0, size)
	scanner := bufio.NewScanner(fi.file)
	for n := 1; scanner.Scan(); n++ {
		select {
		case <-ctx.Done():
			return lines, context.Canceled
		default:
		}

		if start > 0 {
			if n > end {
				break
			}
			if n < start {
				continue
			}
		}

		var line Line
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			return nil, fmt.Errorf("scan line %d failed: %v", n, err)
		}
		lines = append(lines, &line)
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func (fi *fileItem) Write(_ context.Context, b []byte) error {
	fi.Lock()
	defer fi.Unlock()

	if _, err := fi.file.Seek(0, 2); err != nil {
		return err
	}
	if _, err := fi.file.Write(b); err != nil {
		return err
	}
	return nil
}

func NewFile(cfg ConfigFile) (Interface, error) {
	if len(cfg.Dir) == 0 {
		return nil, errors.New("cache dir must be defined")
	}
	if err := os.MkdirAll(cfg.Dir, os.ModePerm); err != nil {
		return nil, err
	}
	return &file{
		dir:     cfg.Dir,
		clients: make(map[string]map[*subscriber]struct{}),
	}, nil
}

type file struct {
	dir     string
	rs      sync.Map
	mux     sync.Mutex
	clients map[string]map[*subscriber]struct{}
}

func (f *file) get(id string) *fileItem {
	v, ok := f.rs.Load(id)
	if !ok {
		return nil
	}
	return v.(*fileItem)
}

func (f *file) read(ctx context.Context, id string, start, size int) ([]*Line, error) {
	fi := f.get(id)
	if fi == nil {
		return nil, nil
	}
	return fi.Read(ctx, start, size)
}

func (f *file) List(ctx context.Context, id string) ([]*Line, error) {
	return f.read(ctx, id, -1, 0)
}

func (f *file) Watch(ctx context.Context, id string) (<-chan *Line, <-chan struct{}, error) {
	f.mux.Lock()
	clients, ok := f.clients[id]
	f.mux.Unlock()
	if !ok {
		return nil, nil, nil
	}

	caches, err := f.List(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	sub := newSubscriber()
	for _, line := range caches {
		sub.publish(line)
	}
	clients[sub] = struct{}{}

	go func() {
		select {
		case <-sub.waitForClose():
		case <-ctx.Done():
			sub.close()
		}
	}()
	return sub.handler, sub.waitForClose(), nil
}

func (f *file) Write(ctx context.Context, id string, line *Line, args ...any) error {
	fi := f.get(id)
	if fi == nil {
		return fmt.Errorf("log stream not found for %s", id)
	}
	b, err := json.Marshal(line)
	if err != nil {
		return fmt.Errorf("log line marshal failed: %v", err)
	}
	b = append(b, []byte("\n")...)

	if err := fi.Write(ctx, b); err != nil {
		return fmt.Errorf("log line write failed: %v", err)
	}
	fi.count++

	f.mux.Lock()
	clients, ok := f.clients[id]
	if !ok {
		f.clients[id] = make(map[*subscriber]struct{})
	}
	f.mux.Unlock()

	for _, arg := range args {
		if v, ok := arg.(PublishOption); ok && !bool(v) {
			return nil
		}
	}
	for client := range clients {
		client.publish(line)
	}
	return nil
}

func (f *file) LineCount(_ context.Context, id string) int {
	fi := f.get(id)
	if fi == nil {
		return 0
	}
	return fi.count
}

func (f *file) Reset(ctx context.Context, id string) error {
	fi := f.get(id)
	if fi == nil {
		return f.Create(ctx, id)
	}
	return fi.file.Truncate(0)
}

func (f *file) Create(_ context.Context, id string) error {
	rwc := f.get(id)
	if rwc != nil {
		return errors.New("log cache file already exist")
	}

	ff, err := os.OpenFile(filepath.Join(f.dir, id), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}

	fi := &fileItem{file: ff}
	f.rs.Store(id, fi)

	f.mux.Lock()
	if _, ok := f.clients[id]; !ok {
		f.clients[id] = make(map[*subscriber]struct{})
	}
	f.mux.Unlock()
	return nil
}

func (f *file) Delete(_ context.Context, id string) error {
	f.mux.Lock()
	clients, ok := f.clients[id]
	f.mux.Unlock()
	if !ok {
		return nil
	}

	// TODO clean clients
	for client := range clients {
		client.close()
	}

	defer f.rs.Delete(id)
	fi := f.get(id)
	if fi != nil {
		if err := fi.Close(); err != nil {
			return err
		}
	}

	f.mux.Lock()
	delete(f.clients, id)
	f.mux.Unlock()
	return nil
}
