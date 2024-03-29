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

package flags

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func NewFlag(name string, value Value, usage string) *flag.Flag {
	return &flag.Flag{
		Name:     name,
		Value:    value,
		DefValue: value.String(),
		Usage:    usage,
	}
}

func NewStringEnvFlag(envPrefix, name, defValue, usage string) *flag.Flag {
	val := GetDefaultEnv(envPrefix, name, defValue)
	value := NewStringValue(val)
	return NewFlag(name, value, usage)
}

func NewStringSliceEnvFlag(envPrefix, name string, defValue []string, usage string) *flag.Flag {
	val := GetDefaultEnv(envPrefix, name, "")
	if val != "" {
		defValue = strings.Split(val, " ")
	}
	value := NewStringSliceValue(defValue)
	return NewFlag(name, value, usage)
}

func NewBoolEnvFlag(envPrefix, name string, defValue bool, usage string) *flag.Flag {
	val := GetDefaultEnv(envPrefix, name, strconv.FormatBool(defValue))
	value := NewBoolValue(false)
	_ = value.Set(val)
	return NewFlag(name, value, usage)
}

type Value interface {
	flag.Value

	Get() any
}

func NewStringValue(val string) Value {
	p := new(string)
	return NewStringVarValue(val, p)
}

func NewStringVarValue(val string, p *string) Value {
	*p = val
	return (*stringValue)(p)
}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

func NewStringSliceValue(val []string) Value {
	p := make([]string, 0)
	return NewStringSliceVarValue(val, &p)
}

func NewStringSliceVarValue(val []string, p *[]string) Value {
	*p = val
	return &stringSlice{value: p}
}

type stringSlice struct {
	value   *[]string
	changed bool
}

func (s *stringSlice) Set(val string) error {
	if s.value == nil {
		return fmt.Errorf("no target (nil pointer to []string)")
	}
	if !s.changed {
		*s.value = make([]string, 0)
	}
	*s.value = append(*s.value, val)
	s.changed = true
	return nil
}

func (s *stringSlice) Get() any {
	return *s.value
}

func (s *stringSlice) String() string {
	if s == nil || s.value == nil {
		return ""
	}
	return "[" + strings.Join(*s.value, " ") + "]"
}

func NewBoolValue(val bool) Value {
	p := new(bool)
	return NewBoolVarValue(val, p)
}

func NewBoolVarValue(val bool, p *bool) Value {
	*p = val
	return (*boolValue)(p)
}

type boolValue bool

func (b *boolValue) Set(s string) error {
	v, _ := strconv.ParseBool(s)
	*b = boolValue(v)
	return nil
}

func (b *boolValue) Get() any { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }
