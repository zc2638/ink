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
)

func New(name string, value Value, usage string) *flag.Flag {
	return &flag.Flag{
		Name:     name,
		Value:    value,
		DefValue: value.String(),
		Usage:    usage,
	}
}

type Value interface {
	flag.Value

	Get() any
}

type stringValue string

func NewStringValue(val string) Value {
	p := new(string)
	return NewStringVarValue(val, p)
}

func NewStringVarValue(val string, p *string) Value {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }
