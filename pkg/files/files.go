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

package files

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrUnknownExt = errors.New("unknown ext")

type Kind string

const (
	KindYAML Kind = "yaml"
	KindJSON Kind = "json"
)

type Item struct {
	Kind Kind
	Data []byte
}

func ReadFile(name string) (*Item, error) {
	item := new(Item)
	ext := strings.TrimPrefix(filepath.Ext(name), ".")
	switch ext {
	case "yaml", "yml":
		item.Kind = KindYAML
	case "json":
		item.Kind = KindJSON
	default:
		return nil, ErrUnknownExt
	}

	fb, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	item.Data = fb
	return item, nil
}

func ReadFiles(fp string) ([]Item, error) {
	if len(fp) == 0 {
		return nil, errors.New("file path is not defined")
	}

	stat, err := os.Stat(fp)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		item, err := ReadFile(fp)
		if err != nil {
			return nil, err
		}
		return []Item{*item}, nil
	}

	dirs, err := os.ReadDir(fp)
	if err != nil {
		return nil, err
	}
	var set []Item
	for _, entry := range dirs {
		if entry.IsDir() {
			continue
		}

		name := filepath.Join(fp, entry.Name())
		item, err := ReadFile(name)
		if errors.Is(err, ErrUnknownExt) {
			continue
		}
		if err != nil {
			return nil, err
		}
		set = append(set, *item)
	}
	return set, nil
}
