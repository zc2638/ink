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

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/zc2638/ink/pkg/api/core/v1"
)

const specFilename = "spec.json"

type ConfigFile struct {
	Dir string `json:"dir"`
}

func NewFile(cfg ConfigFile) (Interface, error) {
	dir := cfg.Dir
	stat, err := os.Stat(dir)
	if err == nil {
		if !stat.IsDir() {
			return nil, fmt.Errorf("path '%s' must be a directory", dir)
		}
	} else {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.MkdirAll(dir, 0644); err != nil {
			return nil, err
		}
	}
	return &file{dir: dir}, nil
}

type file struct {
	dir string
}

func (s *file) List(_ context.Context, meta v1.Metadata, selector Selector) ([]v1.Object, error) {
	dir := filepath.Join(s.dir, meta.Namespace, meta.Kind)
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("path '%s' must be a directory", dir)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []v1.Object
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		name := f.Name()
		fp := filepath.Join(dir, name, specFilename)
		fileBytes, err := os.ReadFile(fp)
		if err != nil {
			return nil, err
		}
		usObj := new(v1.UnstructuredObject)
		if err := usObj.UnmarshalJSON(fileBytes); err != nil {
			return nil, err
		}

		matched := true
		labels := usObj.GetLabels()
		for k, v := range selector {
			if labels[k] != v {
				matched = false
				break
			}
		}
		if matched {
			result = append(result, usObj)
		}
	}
	return result, nil
}

func (s *file) Info(_ context.Context, meta v1.Metadata) (v1.Object, error) {
	fp := filepath.Join(s.dir, meta.Namespace, meta.Kind, meta.Name, specFilename)
	fileBytes, err := os.ReadFile(fp)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	usObj := new(v1.UnstructuredObject)
	if err := usObj.UnmarshalJSON(fileBytes); err != nil {
		return nil, err
	}
	return usObj, nil
}

func (s *file) Create(_ context.Context, meta v1.Metadata, object v1.Object) error {
	stat, err := os.Stat(s.dir)
	if err != nil {
		return fmt.Errorf("stat path '%s' failed: %w", s.dir, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("path '%s' must be a directory", s.dir)
	}

	specPath := filepath.Join(s.dir, meta.Namespace, meta.Kind, meta.Name, specFilename)
	_, err = os.Stat(specPath)
	if err == nil {
		return fmt.Errorf("resource (%s) has already exists", meta.String())
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(specPath)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return err
	}

	b, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return os.WriteFile(specPath, b, 0600)
}

func (s *file) Update(_ context.Context, meta v1.Metadata, object v1.Object) error {
	stat, err := os.Stat(s.dir)
	if err != nil {
		return fmt.Errorf("stat path '%s' failed: %w", s.dir, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("path '%s' must be a directory", s.dir)
	}

	specPath := filepath.Join(s.dir, meta.Namespace, meta.Kind, meta.Name, specFilename)
	_, err = os.Stat(specPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("resource (%s) not exist", meta.String())
	}
	if err != nil {
		return err
	}

	dir := filepath.Dir(specPath)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return err
	}

	b, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return os.WriteFile(specPath, b, 0600)
}

func (s *file) Delete(_ context.Context, meta v1.Metadata) error {
	specPath := filepath.Join(s.dir, meta.Namespace, meta.Kind, meta.Name, specFilename)
	_, err := os.Stat(specPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("resource (%s) not exist", meta.String())
	}
	if err != nil {
		return fmt.Errorf("stat path '%s' failed: %w", specPath, err)
	}
	return os.Remove(filepath.Dir(specPath))
}
